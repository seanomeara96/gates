package repos

import (
	"database/sql"
	"errors"
	"fmt"

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
		return nil, fmt.Errorf("failed to save cart. %v", err)
	}
	return &res, nil
}

func (r *CartRepo) CartExists(id string) (bool, error) {
	var count int
	err := r.db.QueryRow(`SELECT count(id) AS count FROM cart WHERE id = ?`, id).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *CartRepo) GetCartByID(id string) (*models.Cart, error) {
	cart, err := r.selectCart(id)
	if err != nil {
		return nil, err
	}
	if cart.Items, err = r.selectCartItems(cart.ID); err != nil {
		return nil, err
	}
	for i := range cart.Items {
		if cart.Items[i].Components, err = r.selectCartItemComponents(
			cart.ID,
			cart.Items[i].ID,
		); err != nil {
			return nil, err
		}
		cart.Items[i].SetName()
		cart.Items[i].SetPrice()
	}

	cart.SetTotalValue()

	return &cart, nil
}

func (r *CartRepo) selectCart(id string) (models.Cart, error) {
	row := r.db.QueryRow(`
	SELECT
		id,
		created_at,
		last_updated_at,
		total_value
	FROM
		cart
	WHERE
		id = ?`, id)
	var cart models.Cart
	if err := row.Scan(
		&cart.ID,
		&cart.CreatedAt,
		&cart.LastUpdatedAt,
		&cart.TotalValue,
	); err != nil {
		return models.Cart{}, err
	}
	return cart, nil
}

func (r *CartRepo) selectCartItem(cartID, itemID string) (*models.CartItem, error) {
	var ci models.CartItem
	err := r.db.QueryRow(`
	SELECT
		id,
		cart_id,
		name,
		sale_price,
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
		&ci.Name,
		&ci.SalePrice,
		&ci.Qty,
		&ci.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &ci, nil
}

func (r *CartRepo) selectCartItems(cartID string) ([]models.CartItem, error) {
	rows, err := r.db.Query(`
	SELECT
		id,
		cart_id,
		name,
		sale_price,
		qty,
		created_at
	FROM
		cart_item
	WHERE
		cart_id = ?`, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cartItems := []models.CartItem{}
	for rows.Next() {
		var cartItem models.CartItem
		if err := rows.Scan(
			&cartItem.ID,
			&cartItem.CartID,
			&cartItem.Name,
			&cartItem.SalePrice,
			&cartItem.Qty,
			&cartItem.CreatedAt,
		); err != nil {
			return nil, err
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
		return nil, err
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
			return nil, err
		}
		product, err := r.productRepo.GetProductByID(component.Product.Id)
		if err != nil {
			return nil, err
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
		return fmt.Errorf("could not insert cart item in db %w", err)
	}
	return nil
}

func (r *CartRepo) doesCartItemExist(cartID string, cartItemID string) (bool, error) {
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
		return false, fmt.Errorf("could not count cart_item %w", err)
	}
	return count > 0, nil
}

func (r *CartRepo) GetCartByUserID(userID int) (*models.Cart, error) {
	var cart models.Cart
	err := r.db.QueryRow(`
		SELECT
			id, created_at, last_updated_at
		FROM
			carts
		WHERE
			user_id = ?`,
		userID,
	).Scan(
		&cart.ID,
		&cart.CreatedAt,
		&cart.LastUpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *CartRepo) GetCartItemsByCartID(cartID string) ([]*models.CartItem, error) {
	return nil, errors.New("getting cart items not implemented")
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
			c.Qty,
			c.CreatedAt,
		); err != nil {
			return err
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
		return err
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
		return err
	}
	return nil
}

func (r *CartRepo) RemoveCartItem(cartID, itemID string) error {
	return errors.New("remove cart item not implemented")
}

func (r *CartRepo) RemoveCartItemComponents(itemID string) error {
	return errors.New("remove Cart Item Components not yet implemented")
}

/* repository funcs end */
