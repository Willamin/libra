require "stripe"

module StripeHelper
  def all_items
    key = ENV["STRIPE_KEY_SECRET"]
    if key.nil?
      puts("Missing STRIPE_KEY_SECRET, using demo items")
      return demo_items
    end

    Stripe.api_key = key
    skus = Stripe::SKU.list()
    skus.map do |sku|
      product_id = sku.product
      product = Stripe::Product.retrieve(product_id)

      item = Item.new
      item.sku = sku.id
      item.price = sku.price
      item.name = product.name
      item.image = "http://placehold.it/400x200"
      item
    end
  end

  class Item
    attr_accessor :sku, :price, :name, :image
  end

  def demo_items
    [
      Item.new.tap do |i|
        i.sku = "mug"
        i.price = 500
        i.name = "Mug"
        i.image = "http://placehold.it/400x200"
      end,
      Item.new.tap do |i|
        i.sku = "towel"
        i.price = 1000
        i.name = "Towel"
        i.image = "http://placehold.it/400x200"
      end,
    ]
  end
end
