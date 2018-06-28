config[:site_name] = "Libra"
config[:stripe_key_public] = "pk_test_Vo4mlNiNFZ9p265zqws5Xr0i"
config[:copyright] = "Will Lewis"
config[:demo] = true
config[:lambda_url] = "/.netlify/functions/payment"

configure :build do
  config[:host] = "https://libra-shop.org"
end

activate :autoprefixer do |prefix|
  prefix.browsers = "last 2 versions"
end

page '/*.xml', layout: false
page '/*.json', layout: false
page '/*.txt', layout: false

helpers do
  def pricify(cents)
    sprintf("%01.2f", cents / 100.0)
  end

  def readme_contents
    File.read("./README.md").split("---").first
  end
end

require "lib/stripe_helper"
helpers StripeHelper
