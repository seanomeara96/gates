package render

import (
	"fmt"
	"html/template"
	"io"

	"github.com/seanomeara96/gates/models"
)

type Renderer struct {
	tmpl *template.Template
}

func NewRenderer(tmpl *template.Template) *Renderer {
	return &Renderer{
		tmpl,
	}
}

type User struct {
	Email string
}

type BasePageData struct {
	PageTitle       string
	MetaDescription string
	User            User
}

type HomePageData struct {
	FeaturedGates  []*models.Product
	PopularBundles []*models.Product
	BasePageData
}

func (r *Renderer) HomePage(wr io.Writer, data HomePageData) error {
	// validation logic can go here
	return r.tmpl.ExecuteTemplate(wr, "home", data)
}

type ProductPageData struct {
	BasePageData
	Product *models.Product
}

func (r *Renderer) ProductPage(wr io.Writer, data ProductPageData) error {
	return r.tmpl.ExecuteTemplate(wr, "product", data)
}

type BundlePageData struct {
	BasePageData
	Bundle *models.Bundle
}

func (r *Renderer) BundlePage(wr io.Writer, data BundlePageData) error {
	return r.tmpl.ExecuteTemplate(wr, "bundle", data)
}

type ProductsPageData struct {
	Heading string
	BasePageData
	Products []*models.Product
}

func (r *Renderer) ProductsPage(wr io.Writer, data ProductsPageData) error {
	if data.Heading == "" {
		return fmt.Errorf("products page requires a heading, exoected somethig nother than %s", data.Heading)
	}
	return r.tmpl.ExecuteTemplate(wr, "products", data)
}

type WebPageData struct {
	BasePageData
	CustomHTML string
}

func (r *Renderer) WebPage(wr io.Writer, data WebPageData) error {
	return r.tmpl.ExecuteTemplate(wr, "page", data)
}

type NotFoundPageData struct {
	BasePageData
}

func (r *Renderer) NotFoundPage(wr io.Writer, data NotFoundPageData) error {
	return r.tmpl.ExecuteTemplate(wr, "not-found", data)
}

type ProductCardData = models.Product

func (r *Renderer) ProductCard(wr io.Writer, data ProductPageData) error {
	return r.tmpl.ExecuteTemplate(wr, "product-card", data)
}

type BundleBuildResultsData struct {
	RequestedBundleSize float32
	Bundles             []models.Bundle
}

func (r *Renderer) BundleBuildResults(wr io.Writer, data BundleBuildResultsData) error {
	return r.tmpl.ExecuteTemplate(wr, "build-results", data)
}
