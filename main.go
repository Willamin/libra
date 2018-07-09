package main

import (
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/charge"
	"github.com/stripe/stripe-go/product"
	"github.com/stripe/stripe-go/sku"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
)

const (
	AWSLambdaFunctionVersion = "AWS_LAMBDA_FUNCTION_VERSION"
	StripeApiKey             = "STRIPE_KEY_SECRET"
)

type Product struct {
	Name string
	Sku  string
	Cost int64
}

func allProducts() []Product {
	products := make([]Product, 1)

	params := &stripe.SKUListParams{}
	s := sku.List(params)

	for s.Next() {
		s := s.SKU()
		p, err := product.Get(s.Product.ID, nil)
		if err == nil {
			products = append(products, Product{
				Sku:  s.ID,
				Name: p.Name,
				Cost: s.Price,
			})
		}
	}

	return products
}

func (a App) getProductBySku(sku string) Product {
	defaultResponse := allProducts()[0]
	if a.Err != nil {
		return defaultResponse
	}

	for _, v := range allProducts() {
		if v.Sku == sku {
			return v
		}
	}
	a.Err = errors.New(fmt.Sprintf("%s not found", sku))
	return defaultResponse
}

func errorResponse(err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       "Something went wrong",
		StatusCode: 500,
	}, err
}

func (a App) getParamFromBody(body string, param string) string {
	defaultResponse := ""
	if a.Err != nil {
		return defaultResponse
	}
	q, err := url.ParseQuery(body)
	if err != nil {
		a.Err = err
	}
	token := q.Get(param)
	return token
}

func (a App) buildCharge(product Product, token string) {
	if a.Err != nil {
		return
	}

	chargeParams := &stripe.ChargeParams{
		Amount:      stripe.Int64(product.Cost),
		Currency:    stripe.String("usd"),
		Description: stripe.String(fmt.Sprintf("Charge for %s", product.Name)),
	}
	chargeParams.SetSource(token)

	_, err := charge.New(chargeParams)
	if err != nil {
		a.Err = err
	}
}

func simpleResponse() (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "<h1>Thanks for your purchase</h1>",
	}, nil
}

func redirectResponse(newLocation string) (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{
		"Location": newLocation,
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 302,
		Headers:    headers,
	}, nil
}

func (a App) getEnv(keyname string) string {
	defaultResponse := ""
	if a.Err != nil {
		return defaultResponse
	}

	key, found := os.LookupEnv(keyname)
	if found == false {
		a.Err = errors.New(fmt.Sprintf("%s not found", StripeApiKey))
		return defaultResponse
	}
	return key
}

func logRequest(request events.APIGatewayProxyRequest) {
	log.Printf("Handling request:")
	log.Printf("%v", request)
	log.Printf("Params:")
	log.Printf("%v", request.Body)
}

type App struct {
	Err error
}

func appInit() App {
	return App{
		Err: nil,
	}
}

func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	app := appInit()

	key := app.getEnv(StripeApiKey)
	stripe.Key = key

	token := app.getParamFromBody(request.Body, "stripeToken")
	sku := app.getParamFromBody(request.Body, "sku")
	product := app.getProductBySku(sku)
	app.buildCharge(product, token)

	if app.Err != nil {
		return errorResponse(app.Err)
	}

	callback := app.getParamFromBody(request.Body, "callback")
	if app.Err != nil {
		return simpleResponse()
	}

	return redirectResponse(callback)
}

func DoCli() {
	key, found := os.LookupEnv(StripeApiKey)
	stripe.Key = key
	if found == false {
		fmt.Printf("%s not found\n", StripeApiKey)
		os.Exit(1)
	}

	fmt.Printf("%v\n", allProducts())
}

func DemoItems() []Product {
	return []Product{
		Product{
			Sku:  "mug",
			Cost: 500,
			Name: "Mug",
		},
		Product{
			Sku:  "towel",
			Cost: 1000,
			Name: "Towel",
		},
	}
}

func MockHandler() {
	log.Printf("Building Middleman first...")

	exec.Command("middleman", "build").Run()

	http.HandleFunc("/.netlify/functions/payment", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/thanks", 302)
	})

	http.Handle("/", http.FileServer(http.Dir("build")))

	log.Printf("Listening at http://127.0.0.1:4000")
	http.ListenAndServe(":4000", nil)
}

func main() {
	_, ok := os.LookupEnv(AWSLambdaFunctionVersion)
	if ok {
		log.Printf("Running in AWS lambda environment, starting lambda handler.")
		lambda.Start(HandleRequest)
		os.Exit(0)
	}

	log.Printf("Not running in AWS lambda environment, starting mock handler.")
	// DoCli()
	MockHandler()
	os.Exit(0)
}
