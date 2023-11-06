package render

import (
	"fmt"
	"html/template"
	"io"

	"github.com/seanomeara96/gates/config"
	"github.com/seanomeara96/gates/models"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Renderer struct {
	tmpl *template.Template
	env  config.Environment
}

func NewRenderer(pathToTemplates string, env config.Environment) *Renderer {
	funcMap := template.FuncMap{
		"sizeRange": func(width, tolerance float32) float32 {
			return width - tolerance
		},
		"title": func(str string) string {
			return cases.Title(language.AmericanEnglish).String(str)
		},
	}

	tmpl := template.New("gate-builder").Funcs(funcMap)
	tmpl = template.Must(tmpl.ParseGlob(pathToTemplates))
	return &Renderer{
		tmpl,
		env,
	}
}

type templateData struct {
	Env config.Environment
}

func (r *Renderer) NewTemplateData() templateData {
	return templateData{
		Env: r.env,
	}
}

type basePageData struct {
	templateData
	PageTitle       string
	MetaDescription string
	User            models.User
}

func (r *Renderer) NewBasePageData(pageTitle string, metaDescription string, user models.User) basePageData {
	templateData := r.NewTemplateData()
	return basePageData{
		templateData:    templateData,
		PageTitle:       pageTitle,
		MetaDescription: metaDescription,
		User:            user,
	}
}

type homePageData struct {
	FeaturedGates  []*models.Product
	PopularBundles []*models.Product
	basePageData
}

func (r *Renderer) NewHomePageData(
	featuredGates []*models.Product,
	popularBundles []*models.Product,
	basePageData basePageData,
) homePageData {
	return homePageData{
		FeaturedGates:  featuredGates,
		PopularBundles: popularBundles,
		basePageData:   basePageData,
	}
}

func (r *Renderer) HomePage(wr io.Writer, data homePageData) error {
	// validation logic can go here
	return r.tmpl.ExecuteTemplate(wr, "home", data)
}

type productPageData struct {
	BasePageData basePageData
	Product      *models.Product
}

func (r *Renderer) NewProductPageData(basePageData basePageData, product *models.Product) productPageData {
	return productPageData{
		BasePageData: basePageData,
		Product:      product,
	}
}

func (r *Renderer) ProductPage(wr io.Writer, data productPageData) error {
	return r.tmpl.ExecuteTemplate(wr, "product", data)
}

type bundlePageData struct {
	BasePageData basePageData
	Bundle       *models.Bundle
}

func (r *Renderer) NewBundlePageData(basePageData basePageData, bundle *models.Bundle) bundlePageData {
	return bundlePageData{
		BasePageData: basePageData,
		Bundle:       bundle,
	}
}

func (r *Renderer) BundlePage(wr io.Writer, data bundlePageData) error {
	return r.tmpl.ExecuteTemplate(wr, "bundle", data)
}

type productsPageData struct {
	Heading      string
	BasePageData basePageData
	Products     []*models.Product
}

func (r *Renderer) NewProductsPageData(basePageData basePageData, heading string, products []*models.Product) productsPageData {
	return productsPageData{
		Heading:      heading,
		BasePageData: basePageData,
		Products:     products,
	}
}

func (r *Renderer) ProductsPage(wr io.Writer, data productsPageData) error {
	if data.Heading == "" {
		return fmt.Errorf("products page requires a heading, exoected somethig nother than %s", data.Heading)
	}
	return r.tmpl.ExecuteTemplate(wr, "products", data)
}

type cartPageData struct {
	BasePageData basePageData
	Cart         *models.Cart
	CartItems    []*models.CartItem
}

func (r *Renderer) NewCartPageData(basePageData basePageData, cart *models.Cart, cartItems []*models.CartItem) cartPageData {
	return cartPageData{
		BasePageData: basePageData,
		Cart:         cart,
		CartItems:    cartItems,
	}
}

func (r *Renderer) CartPage(wr io.Writer, cartPageData cartPageData) error {
	return r.tmpl.ExecuteTemplate(wr, "cart", cartPageData)
}

type webPageData struct {
	BasePageData basePageData
	CustomHTML   string
}

func (r *Renderer) NewWebPageData(basePageData basePageData, customHTML string) webPageData {
	return webPageData{
		BasePageData: basePageData,
		CustomHTML:   customHTML,
	}
}

func (r *Renderer) WebPage(wr io.Writer, data webPageData) error {
	return r.tmpl.ExecuteTemplate(wr, "page", data)
}

type notFoundPageData struct {
	BasePageData basePageData
}

func (r *Renderer) NotFoundPageData(basePageData basePageData) notFoundPageData {
	return notFoundPageData{BasePageData: basePageData}
}

func (r *Renderer) NotFoundPage(wr io.Writer, data notFoundPageData) error {
	return r.tmpl.ExecuteTemplate(wr, "not-found", data)
}

type productCardData struct {
	TemplateData templateData
	Product      models.Product
}

func (r *Renderer) NewProductCardData(templateData templateData, product models.Product) productCardData {
	return productCardData{
		TemplateData: r.NewTemplateData(),
		Product:      product,
	}
}

func (r *Renderer) ProductCard(wr io.Writer, data productCardData) error {
	return r.tmpl.ExecuteTemplate(wr, "product-card", data)
}

type bundleBuildResultsData struct {
	TemplateData        templateData
	RequestedBundleSize float32
	Bundles             []models.Bundle
}

func (r *Renderer) NewBundleBuildResultsData(
	requestedBundleSize float32,
	bundles []models.Bundle,
) bundleBuildResultsData {
	return bundleBuildResultsData{
		TemplateData:        r.NewTemplateData(),
		RequestedBundleSize: requestedBundleSize,
		Bundles:             bundles,
	}
}

func (r *Renderer) BundleBuildResults(wr io.Writer, data bundleBuildResultsData) error {
	return r.tmpl.ExecuteTemplate(wr, "build-results", data)
}
