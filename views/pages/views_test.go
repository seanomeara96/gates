package pages

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

import "github.com/a-h/templ"

func TestRender(t *testing.T) {

	var tests = []func() templ.Component{
		func() templ.Component {
			props := BaseProps{}
			return Base(props)
		},
		func() templ.Component {
			props := HomeProps{}
			return Home(props)
		},
		func() templ.Component {
			props := ProductsPageProps{}
			return Products(props)
		},
		func() templ.Component {
			props := ContactPageProps{}
			return Contact(props)
		},
		func() templ.Component {
			props := TestPageProps{}
			return Test(props)
		},
		func() templ.Component {
			props := OrderSuccessPageProps{}
			return OrderSuccess(props)
		},
		func() templ.Component {
			props := ProductPageProps{}
			return Product(props)
		},
		func() templ.Component {
			props := NotFoundPageProps{}
			return NotFound(props)
		},
		func() templ.Component {
			props := InternalErrorPageProps{}
			return SomethingWentWrong(props)
		},
		func() templ.Component {
			props := CartPageProps{}
			return Cart(props)
		},
		func() templ.Component {
			props := InternalErrorPageProps{}
			return InternalError(props)
		},
		func() templ.Component {
			props := DashboardPageProps{}
			return Dashboard(props)
		},
		func() templ.Component {
			props := AdminLoginPageProps{}
			return AdminLogin(props)
		},
		func() templ.Component {
			props := AdminPageProps{}
			return Admin(props)
		},
	}

	for _, test := range tests {
		component := test()
		err := component.Render(context.Background(), io.Discard)
		require.NoError(t, err)
	}

}
