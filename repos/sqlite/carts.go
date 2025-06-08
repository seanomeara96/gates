package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/seanomeara96/gates/models"
)

type productRepo interface {
	GetProductByID(id int) (*models.Product, error)
}

type CartRepo struct {
	db          *sql.DB
	productRepo productRepo
}

func NewCartRepo(db *sql.DB, productRepo productRepo) *CartRepo {
	if db == nil {
		// Consider panic or returning an error if a nil db is critical
		panic("database connection is nil for CartRepo")
	}
	if productRepo == nil {
		panic("product repo is nil for cart repo")
	}
	return &CartRepo{db, productRepo}
}

func (r *CartRepo) SaveCart(cart models.Cart) (*sql.Result, error) {
	res, err := r.db.Exec(`INSERT INTO
		cart(
			id,
			created_at,
			last_updated_at)
		VALUES
			(?, ?, ?)`,
		cart.ID,
		cart.CreatedAt,
		cart.LastUpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save cart: %v", err)
	}
	return &res, nil
}

func (r *CartRepo) CartExists(id string) (bool, error) {
	var count int
	err := r.db.QueryRow(`SELECT count(id) AS count FROM cart WHERE id = ?`, id).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if cart exists: %v", err)
	}
	return count > 0, nil
}

func (r *CartRepo) GetCartByID(id string) (*models.Cart, bool, error) {
	cart, found, err := r.selectCart(id)
	if err != nil {
		return nil, found, fmt.Errorf("failed to select cart with ID %s: %v", id, err)
	}

	if !found {
		return nil, found, nil
	}

	if cart.Items, err = r.selectCartItems(cart.ID); err != nil {
		return nil, found, fmt.Errorf("failed to select items for cart %s: %v", cart.ID, err)
	}
	for i := range cart.Items {
		if cart.Items[i].Components, err = r.selectCartItemComponents(
			cart.ID,
			cart.Items[i].ID,
		); err != nil {
			return nil, found, fmt.Errorf("failed to select components for cart item %s: %v", cart.Items[i].ID, err)
		}
		cart.Items[i].SetName()
		cart.Items[i].SetPrice()
	}

	cart.SetTotalValue()

	return &cart, found, nil
}

func (r *CartRepo) selectCart(id string) (models.Cart, bool, error) {
	row := r.db.QueryRow(`
	SELECT
		id,
		created_at,
		last_updated_at
	FROM
		cart
	WHERE
		id = ?`, id)
	var cart models.Cart
	if err := row.Scan(
		&cart.ID,
		&cart.CreatedAt,
		&cart.LastUpdatedAt,
	); err != nil {

		if err == sql.ErrNoRows {
			return models.Cart{}, false, nil
		}

		return models.Cart{}, true, fmt.Errorf("failed to scan cart data: %v", err)
	}
	return cart, true, nil
}

func (r *CartRepo) SelectCartItem(cartID, itemID string) (*models.CartItem, error) {
	var ci models.CartItem
	err := r.db.QueryRow(`
	SELECT
		id,
		cart_id,
		qty,
		created_at
	FROM
		cart_item
	WHERE
		cart_id = ?
	AND
		id = ?
	`,
		cartID,
		itemID,
	).Scan(
		&ci.ID,
		&ci.CartID,
		&ci.Qty,
		&ci.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to select cart item (cartID: %s, itemID: %s): %v", cartID, itemID, err)
	}
	return &ci, nil
}

func (r *CartRepo) selectCartItems(cartID string) ([]models.CartItem, error) {
	rows, err := r.db.Query(`
	SELECT
		id,
		cart_id,
		qty,
		created_at
	FROM
		cart_item
	WHERE
		cart_id = ?`, cartID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cart items for cart %s: %v", cartID, err)
	}
	defer rows.Close()
	cartItems := []models.CartItem{}
	for rows.Next() {
		var cartItem models.CartItem
		if err := rows.Scan(
			&cartItem.ID,
			&cartItem.CartID,
			&cartItem.Qty,
			&cartItem.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan cart item: %v", err)
		}
		cartItems = append(cartItems, cartItem)
	}
	return cartItems, nil
}

func (r *CartRepo) selectCartItemComponents(cartID, cartItemID string) ([]models.CartItemComponent, error) {
	rows, err := r.db.Query(`
	SELECT
		cart_item_id,
		cart_id,
		product_id,
		qty,
		created_at
	FROM
		cart_item_component
	WHERE
		cart_item_id = ?
	AND
		cart_id = ?`,
		cartItemID, cartID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cart item components (cartID: %s, itemID: %s): %v", cartID, cartItemID, err)
	}
	defer rows.Close()

	components := []models.CartItemComponent{}
	for rows.Next() {
		var component models.CartItemComponent
		if err := rows.Scan(
			&component.CartItemID,
			&component.CartID,
			&component.Product.Id,
			&component.Product.Qty,
			&component.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan cart item component: %v", err)
		}
		product, err := r.productRepo.GetProductByID(component.Product.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to get product (ID: %d) for cart component: %v", component.Product.Id, err)
		}
		product.Qty = component.Product.Qty
		component.Product = *product
		components = append(components, component)
	}
	return components, nil
}

func (r *CartRepo) InsertCartItem(cartItem models.CartItem) error {
	q := `
	INSERT INTO
		cart_item (
			id,
			cart_id,
			qty,
			created_at
		)
	VALUES
		(?, ?, ?, ?)`
	_, err := r.db.Exec(
		q,
		cartItem.ID,
		cartItem.CartID,
		cartItem.Qty,
		cartItem.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("could not insert cart item in db (ID: %s, cartID: %s): %w", cartItem.ID, cartItem.CartID, err)
	}
	return nil
}

/*
Important to keep track of this for the purposes of abandoned cart messages
*/
func (r *CartRepo) SetLastUpdated(cartID string) error {
	if _, err := r.db.Exec("UPDATE cart SET last_updated_at = ? WHERE id = ?", time.Now(), cartID); err != nil {
		return fmt.Errorf("could not update last_updated_at on cart (ID: %s): %w", cartID, err)
	}
	return nil
}

func (r *CartRepo) DoesCartItemExist(cartID string, cartItemID string) (bool, error) {
	var count int
	err := r.db.QueryRow(`
	SELECT
		count(id) as count
	FROM
		cart_item
	WHERE
		id = ?
	AND
		cart_id = ?`,
		cartItemID, cartID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("could not check if cart item exists (cartID: %s, itemID: %s): %w", cartID, cartItemID, err)
	}
	return count > 0, nil
}

func (r *CartRepo) GetCartByUserID(userID int) (*models.Cart, error) {
	var cart models.Cart
	err := r.db.QueryRow(`
		SELECT
			id, created_at, last_updated_at
		FROM
			cart
		WHERE
			user_id = ?`,
		userID,
	).Scan(
		&cart.ID,
		&cart.CreatedAt,
		&cart.LastUpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart for user (ID: %d): %v", userID, err)
	}
	return &cart, nil
}

func (r *CartRepo) GetCartItemsByCartID(cartID string) ([]*models.CartItem, error) {
	return nil, fmt.Errorf("getting cart items not implemented for cartID: %s", cartID)
}

func (r *CartRepo) SaveCartItemComponents(components []models.CartItemComponent) error {
	for _, c := range components {
		q := `
		INSERT INTO
			cart_item_component (
				cart_item_id,
				cart_id,
				product_id,
				qty,
				created_at
			)
		VALUES
			(?, ?, ?, ?, ?)`
		if _, err := r.db.Exec(q,
			c.CartItemID,
			c.CartID,
			c.Product.Id,
			c.Product.Qty,
			c.CreatedAt,
		); err != nil {
			return fmt.Errorf("failed to save cart item component (cartItemID: %s, productID: %d): %v",
				c.CartItemID, c.Product.Id, err)
		}
	}
	return nil
}

func (r *CartRepo) IncrementCartItem(cartID, itemID string) error {
	if _, err := r.db.Exec(`
		UPDATE
			cart_item
		SET
			qty = qty + 1
		WHERE
			id = ?
		AND
			cart_id = ?`,
		itemID,
		cartID,
	); err != nil {
		return fmt.Errorf("failed to increment cart item (cartID: %s, itemID: %s): %v", cartID, itemID, err)
	}
	return nil
}

func (r *CartRepo) DecrementCartItem(cartID, itemID string) error {
	if _, err := r.db.Exec(`
	UPDATE cart_item
	SET qty = qty - 1
	WHERE id = ?
	AND cart_id = ?`,
		itemID,
		cartID,
	); err != nil {
		return fmt.Errorf("failed to decrement cart item (cartID: %s, itemID: %s): %v", cartID, itemID, err)
	}
	return nil
}

func (r *CartRepo) RemoveCartItem(cartID, itemID string) error {
	return fmt.Errorf("remove cart item not implemented (cartID: %s, itemID: %s)", cartID, itemID)
}

func (r *CartRepo) RemoveCartItemComponents(itemID string) error {
	return fmt.Errorf("remove Cart Item Components not yet implemented for itemID: %s", itemID)
}

/* repository funcs end */
