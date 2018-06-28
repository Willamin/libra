package main

import (
  "errors"
  "fmt"
  "github.com/aws/aws-lambda-go/events"
  "github.com/aws/aws-lambda-go/lambda"
  "github.com/stripe/stripe-go"
  "github.com/stripe/stripe-go/charge"
  "github.com/stripe/stripe-go/sku"
  "github.com/stripe/stripe-go/product"
  "log"
  "net/url"
  "net/http"
  "os"
  "os/exec"
)

const (
  AWSLambdaFunctionVersion = "AWS_LAMBDA_FUNCTION_VERSION"
  StripeApiKey             = "STRIPE_KEY_SECRET"
)

type Product struct {
  Name string
  Sku string
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
        Sku: s.ID,
        Name: p.Name,
        Cost: s.Price,
      })
    }
  }

  return products
}

func getProductBySku(sku string) (Product, error) {
  for _, v := range allProducts() {
    if v.Sku == sku {
      return v, nil
    }
  }
  return allProducts()[0], errors.New(fmt.Sprintf("%s not found", sku))
}

func errorResponse(err error) (events.APIGatewayProxyResponse, error) {
  return events.APIGatewayProxyResponse{
    Body:       "Something went wrong",
    StatusCode: 500,
  }, err
}

func getTokenFromBody(body string) (string, error) {
  q, err := url.ParseQuery(body)

  token := q.Get("stripeToken")
  return token, err
}

func getSkuFromBody(body string) (string, error) {
  q, err := url.ParseQuery(body)

  sku := q.Get("sku")
  return sku, err
}

func getCallbackFromBody(body string) (string, error) {
  q, err := url.ParseQuery(body)

  cb := q.Get("callback")
  return cb, err
}

func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
  var err error = nil

  log.Printf("Handling request:")
  log.Printf("%v", request)
  log.Printf("Params:")
  log.Printf("%v", request.Body)

  key, found := os.LookupEnv(StripeApiKey)
  stripe.Key = key
  if found == false {
    err = errors.New(fmt.Sprintf("%s not found", StripeApiKey))
    return errorResponse(err)
  }

  token, err := getTokenFromBody(request.Body)
  if err != nil {
    return errorResponse(err)
  }

  sku, err := getSkuFromBody(request.Body)
  product, err := getProductBySku(sku)
  if err != nil {
    return errorResponse(err)
  }

  chargeParams := &stripe.ChargeParams{
    Amount:      stripe.Int64(product.Cost),
    Currency:    stripe.String("usd"),
    Description: stripe.String(fmt.Sprintf("Charge for %s", product.Name)),
  }
  chargeParams.SetSource(token)

  _, err = charge.New(chargeParams)
  if err != nil {
    return errorResponse(err)
  }

  callback, err := getCallbackFromBody(request.Body)
  if err != nil {
    return events.APIGatewayProxyResponse{
      StatusCode: 200,
      Body: "<h1>Thanks for your purchase</h1>",
    }, nil
  }

  headers := map[string]string{
    "Location": callback,
  }

  return events.APIGatewayProxyResponse{
    StatusCode: 302,
    Headers:    headers,
  }, nil
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
      Sku: "mug",
      Cost: 500,
      Name: "Mug",
    },
    Product{
      Sku: "towel",
      Cost: 1000,
      Name: "Towel",
    },
  }
}

func MockHandler() {
  log.Printf("Building Middleman first...")

  exec.Command("middleman", "build").Run()

  http.HandleFunc("/.netlify/functions/payment", func (w http.ResponseWriter, r *http.Request) {
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
