# RestockBot

Year 2020 has been quite hard for hardware supply. Graphics cards are out of stock everywhere. Nobody can grab the new generation (AMD RX 6000 series, NVIDIA GeForce RTX 3000 series). Even older generations are hard to find. `RestockBot` is a bot that crawl retailers websites and notify when a product is available.

## Requirements

### Headless browser

Use Docker:

```
docker run --name chromium --rm -d -p 9222:9222 montferret/chromium
```

Or get inspired by the [source code](https://github.com/MontFerret/chromium) to run it on your own.

### Amazon (optional)

To access the [Product Advertising API](https://webservices.amazon.com/paapi5/documentation/) and start to notify for Amazon products, you will need to have a valid [Amazon Associates](https://affiliate-program.amazon.com) account in the [Marketplace](https://github.com/spiegel-im-spiegel/pa-api/blob/v0.9.0/marketplace.go#L36) of your choice. You will then be able to retreive your **partner tag**, and the **Marketplace name** obviously.

Once your account has been validated, you can request access to the Product Advertising API (PA API) to retreive your **access key** and your **secret key**.

Ensure you follow the **terms of services** before subscribing to the Amazon Associates program and use the PA API.

### Twitter (optional)

Follow [this procedure](https://github.com/jouir/twitter-login) to generate all the required settings:
* `consumer_key`
* `consumer_secret`
* `access_token`
* `access_token_secret`

### Telegram (optional)

Follow [this procedure](https://core.telegram.org/bots#3-how-do-i-create-a-bot) to create a bot `token`.

Then you have two possible destinations to send messages:
* channel using a `channel_name` (string)
* chat using a `chat_id` (integer)

For testing purpose, you should store the token in a variable for next sections:
```
read -s TOKEN
```

#### Chat

To get the chat identifier, you can send a message to your bot then read messages using the API:

```
curl -s -XGET "https://api.telegram.org/bot${TOKEN}/getUpdates" | jq -r ".result[].message.chat.id"
```

You can test to send messages to a chat with:

```
read CHAT_ID
curl -s -XGET "https://api.telegram.org/bot${TOKEN}/sendMessage?chat_id=${CHAT_ID}&text=hello" | jq
```

#### Channel

Public channel names can be used (example: `@mychannel`). For private channels, you should use a `chat_id` instead.

You can test to send messages to a channel with:

```
read CHANNEL_NAME
curl -s -XGET "https://api.telegram.org/bot${TOKEN}/sendMessage?chat_id=${CHANNEL_NAME}&text=hello" | jq
```

Don't forget to prefix the channel name with an `@`.

## Compilation

### With pre-built binaries

Download the latest [release](https://github.com/jouir/restockbot/releases).

Ensure checksums are identical.

### With make

Clone the repository:
```
git clone https://github.com/jouir/restockbot.git
```

Build the `restockbot` binary:
```
make build
ls -l bin/restockbot
```

Build with the architecture in the binary name:

```
make release
```

Eventually remove produced binaries with:

```
make clean
```

### With Docker

```
docker image build -t restockbot:$(cat VERSION) .
```

## Configuration

Default file is `restockbot.json` in the current directory. The file name can be passed with the `-config` argument.

Options:

* `urls` (optional): list of retailers web pages
* `amazon` (optional)
    * `searches`: list of keywords to search for (ex: `["nvidia rtx", "amd rx"]`)
    * `access_key`: access key to access the [Product Advertising API](https://webservices.amazon.com/paapi5/documentation/)
    * `secret_key`: secret key to access the [Product Advertising API](https://webservices.amazon.com/paapi5/documentation/)
    * `marketplaces`: list of documents containing a Marketplace `name` and a `partner_tag` (ex: `{"marketplaces":[{"name": "www.amazon.com", "partner_tag": "mytag-01"}]}`)
    * `amazon_fulfilled`: include only products packaged by Amazon
    * `amazon_merchant`: include only products sold by Amazon
    * `affiliate_links`: generate affiliate links with the partner tag
* `twitter` (optional):
    * `consumer_key`: API key of your Twitter application
    * `consumer_secret`: API secret of your Twitter application
    * `access_token`: authentication token generated for your Twitter account
    * `access_token_secret`: authentication token secret generated for your Twitter account
    * `hashtags`: list of key/value used to append hashtags to each tweet. Key is the pattern to match in the product name, value is the string to append to the tweet. For example, `{"twitter": {"hashtags": [{"rtx 3090": "#nvidia #rtx3090"}]}}` will detect `rtx 3090` to append `#nvidia #rtx3090` at the end of the tweet.
* `telegram` (optional):
    * `channel_name`: send message to a channel (ex: `@channel`)
    * `chat_id`: send message to a chat (ex: `1234`)
    * `token`: key returned by BotFather
* `include_regex` (optional): include products with a name matching this regexp
* `exclude_regex` (optional): exclude products with a name matching this regexp
* `browser_address` (optional): set headless browser address (ex: `http://127.0.0.1:9222`)
* `api` (optional):
    * `address`: listen address for the REST API (ex: `127.0.0.1:8000`)
    * `cert_file` (optional): use SSL and use this certificate file
    * `key_file` (optional): use SSL and use this key file

## Usage

### With binary

```
restockbot -help
```

### With Docker

```
docker run -it --name restockbot --rm --link chromium:chromium -v $(pwd):/root/ restockbot:$(cat VERSION) restockbot -help
```

## Execution modes

There are two modes:
* **default**: without special argument, the bot parses websites and manage its own database
* **API**: using the `-api` argument, the bot starts the HTTP API to expose data from the database

## How to contribute

Lint the code with pre-commit:

```
docker run -it -v $(pwd):/mnt/ --rm golang:latest bash
go get -u golang.org/x/lint/golint
apt-get update && apt-get upgrade -y && apt-get install -y git python3-pip
pip3 install pre-commit
cd /mnt
pre-commit run --all-files
```

## How to parse a shop

### Create the Ferret query

`RestockBot` uses [Ferret](https://github.com/MontFerret/ferret) and its FQL (Ferret Query Language) to parse websites. The full documentation is available [here](https://www.montferret.dev/docs/introduction/). Once installed, this library can be used as a CLI command or embedded in the application. To create the query, we can use the CLI for fast iterations, then we'll integrate the query in `RestockBot` later.

```
vim shop.fql
ferret --cdp http://127.0.0.1:9222 -time shop.fql
```

The query must return a list of products in JSON format with the following elements:
* `name`: string
* `url`: string
* `price`: float
* `price_currency`: string
* `available`: boolean

Example:

```json
[
  {
    "available": false,
    "name": "Zotac GeForce RTX 3070 AMP Holo",
    "price": 799.99,
    "price_currency": "EUR",
    "url": "https://www.topachat.com/pages/detail2_cat_est_micro_puis_rubrique_est_wgfx_pcie_puis_ref_est_in20007322.html"
  },
  {
    "available": false,
    "name": "Asus GeForce RTX 3070 DUAL 8G",
    "price": 739.99,
    "price_currency": "EUR",
    "url": "https://www.topachat.com/pages/detail2_cat_est_micro_puis_rubrique_est_wgfx_pcie_puis_ref_est_in20005540.html"
  },
  {
    "available": false,
    "name": "Palit GeForce RTX 3070 GamingPro OC",
    "price": 819.99,
    "price_currency": "EUR",
    "url": "https://www.topachat.com/pages/detail2_cat_est_micro_puis_rubrique_est_wgfx_pcie_puis_ref_est_in20005819.html"
  }
]
```

`RestockBot` will convert this JSON to a list of `Product`.

### Embed the query

Shops are configured as a list of URLs:

```json
{
    "urls": [
        "https://www.topachat.com/pages/produits_cat_est_micro_puis_rubrique_est_wgfx_pcie_puis_f_est_58-11447,11445,11446,11559,11558.html",
        "https://www.ldlc.com/informatique/pieces-informatique/carte-graphique-interne/c4684/+fv121-19183,19184,19185,19339,19340.html",
        "https://www.materiel.net/carte-graphique/l426/+fv121-19183,19184,19185,19339,19340/"
    ]
}
```

The `Parse` function ([parser.go](parser.go)) will be called. In this example, the following **shop names** will be deduced: `topachat.com`, `ldlc.com` and `materiel.net`.

Each shop should implement a function to create a ferret query based on an URL:
* `func createQueryForLDLC(url string) string`
* `func createQueryForMaterielNet(url string) string`
* `func createQueryForTopachat(url string) string`
* ...

This function should be added to the switch of the `createQuery` function ([parser.go](parser.go)).

Products will then be parsed.

## Disclaimer

Crawling a website should be used with caution. Please check with retailers if the bot respects the terms of use for their websites. Authors of the bot are not responsible of the bot usage.
