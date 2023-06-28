
to make a purchase possible i need to send the following info to the server:
- product id
- qty

example: [{ id: 1, qty: 1}, { id: 2, qty: 1}]


for the cart to function as normally id need:
- product name
- product url
- product id
- product qty

example: { 
    productName: "product name",
    productURL: "/product-url",
    productId: 1,
    productPrice: 10,
    productQty: 1,
}


but i need this to play nice with bundles

so cart data ought to look more like this

example: {
    productName: "product name",
    productURL: "/product-url",
    meta: [
        {
            productId: 1,
            productPrice: 10,
            productQty: 1
        },
        {
            productId: 2,
            productPrice: 20,
            productQty: 2
        }
    ]
}

which means data sent to checkout endpoint will look like this if someone buys one product and one bundle

example: [
    [{id: 1, qty: 1}],
    [
        {id: 1, qty: 1},
        {id: 2, qty: 2}
    ]

]


