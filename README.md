# Libra

_An e-commerce platform leveraging the power of static sites and FaaS_

Libra is built using a few technologies/services:

- Middleman generates the site
- Netlify hosts it
- AWS Lambda (via Netlify) verifies transactions
- Stripe processes payments

---

## Demo

A demo site can be found running at [https://libra-shop.org](https://libra-shop.org).

## Usage
To use Libra for your own store:

### Stripe Setup
- create a Stripe account
- get your API keys (public and private)
- add products+skus in Stripe

### Repository/Codebase Setup
- fork/clone this repository
- style files in the `source` directory :art:
- update `demo`, `site_name`, `stripe_key_public`, `copyright`, and `host` in `config.rb`

### Netlify Setup
- connect your repository to Netlify
- add a build environment variable to Netlify for `STRIPE_KEY_SECRET` with the value of your private Stripe API key
- update your domain information on Netlify and your DNS host
- update your `netlify.toml` GO_IMPORT_PATH key
- update the `_redirects` file appropriately
