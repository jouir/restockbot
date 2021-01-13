import logging
from html.parser import HTMLParser

from bs4 import BeautifulSoup
from bs4.element import Tag
from db import Product
from utils import parse_base_url

logger = logging.getLogger(__name__)


# Parsers definitively need to be replaced by beautifulsoup because the code is not maintainable


class ProductParser(HTMLParser):
    def __init__(self):
        super().__init__()
        self.products = []
        self.next_page = None


class TopAchatParser(ProductParser):
    def __init__(self, url=None):
        super().__init__()
        self._parsing_article = False
        self._parsing_availability = False
        self._parsing_price = False
        self._parsing_price_currency = False
        self._parsing_name = False
        self._parsing_url = False
        self._product = Product()
        if url:
            self._base_url = parse_base_url(url)
        else:
            self._base_url = 'https://www.topachat.com'

    @staticmethod
    def parse_name(data):
        return data.split(' + ')[0].strip()

    def handle_starttag(self, tag, attrs):
        if tag == 'article':
            for name, value in attrs:
                if 'grille-produit' in value.split(' '):
                    self._parsing_article = True
        elif self._parsing_article:
            if tag == 'link':
                for name, value in attrs:
                    if name == 'itemprop' and value == 'availability':
                        self._parsing_availability = True
                    elif self._parsing_availability and name == 'href':
                        self._product.available = value != 'http://schema.org/OutOfStock'
            elif tag == 'div':
                for name, value in attrs:
                    if name == 'itemprop' and value == 'price':
                        self._parsing_price = True
                    elif self._parsing_price and name == 'content':
                        self._product.price = float(value)
                    elif name == 'class' and value == 'libelle':
                        self._parsing_url = True
                        self._parsing_name = True
            elif tag == 'meta':
                for name, value in attrs:
                    if name == 'itemprop' and value == 'priceCurrency':
                        self._parsing_price_currency = True
                    elif self._parsing_price_currency and name == 'content':
                        self._product.price_currency = value
            elif tag == 'a':
                for name, value in attrs:
                    if self._parsing_url and name == 'href':
                        self._product.url = f'{self._base_url}{value}'

    def handle_data(self, data):
        if self._parsing_name and self.get_starttag_text().startswith('<h3>') and not self._product.name:
            self._product.name = self.parse_name(data)
            self._parsing_name = False

    def handle_endtag(self, tag):
        if self._parsing_article and tag == 'article':
            self._parsing_article = False
            self.products.append(self._product)
            self._product = Product()
        elif self._parsing_availability and tag == 'link':
            self._parsing_availability = False
        elif self._parsing_price and tag == 'div':
            self._parsing_price = False
        elif self._parsing_price_currency and tag == 'meta':
            self._parsing_price_currency = False


class LDLCParser(ProductParser):
    def __init__(self, url=None):
        super().__init__()
        self._product = Product()
        self.__parsing_pdt_item = False
        self.__parsing_pdt_id = False
        self._parsing_title = False
        self.__parsing_pagination = False
        self.__parsing_next_page_section = False
        self._parsing_stock = False
        self._parsing_price = False
        if url:
            self._base_url = parse_base_url(url)
        else:
            self._base_url = 'https://www.ldlc.com'

    @property
    def _parsing_item(self):
        return self.__parsing_pdt_item and self.__parsing_pdt_id

    @property
    def _parsing_next_page(self):
        return self.__parsing_pagination and self.__parsing_next_page_section

    @staticmethod
    def parse_price(string):
        currency = None
        if '€' in string:
            currency = 'EUR'
        price = int(''.join([i for i in string if i.isdigit()]))
        return price, currency

    def handle_starttag(self, tag, attrs):
        if not self._parsing_item and tag == 'li' and not self.__parsing_pagination:
            for name, value in attrs:
                if name == 'class' and value == 'pdt-item':
                    self.__parsing_pdt_item = True
                elif name == 'id' and value.startswith('pdt-'):
                    self.__parsing_pdt_id = True
        elif not self.__parsing_pagination and tag == 'ul':
            for name, value in attrs:
                if name == 'class' and value == 'pagination':
                    self.__parsing_pagination = True
        elif self.__parsing_pagination and tag == 'li':
            for name, value in attrs:
                if name == 'class' and value == 'next':
                    self.__parsing_next_page_section = True
        elif self._parsing_next_page and tag == 'a':
            for name, value in attrs:
                if name == 'href':
                    self.next_page = f'{self._base_url}{value}'
        elif self._parsing_item:
            if tag == 'h3':
                self._parsing_title = True
            elif self._parsing_title and tag == 'a':
                for name, value in attrs:
                    if name == 'href':
                        self._product.url = f'{self._base_url}{value}'
            elif tag == 'div':
                for name, value in attrs:
                    if not self._parsing_stock and name == 'class' and 'modal-stock-web' in value.split(' '):
                        self._parsing_stock = True
                    elif not self._parsing_price and name == 'class' and value == 'price':
                        self._parsing_price = True

    def handle_data(self, data):
        last_tag = self.get_starttag_text()
        if self._parsing_title and not self._product.name and last_tag.startswith('<a'):
            self._product.name = data.strip()
        elif self._parsing_stock and self._product.available is None and last_tag.startswith('<span>'):
            self._product.available = data.strip() != 'Rupture'
        elif self._parsing_price:
            if last_tag.startswith('<div'):
                self._product.price, self._product.price_currency = self.parse_price(data)
            elif last_tag.startswith('<sup>'):
                self._product.price += int(data) / 100

    def handle_endtag(self, tag):
        if self._parsing_item and tag == 'li':
            self.__parsing_pdt_item = False
            self.__parsing_pdt_id = False
            self.products.append(self._product)
            self._product = Product()
        elif self._parsing_title and tag == 'h3':
            self._parsing_title = False
        elif self._parsing_stock and tag == 'span':
            self._parsing_stock = False
        elif self._parsing_price and tag == 'div':
            self._parsing_price = False
        elif self.__parsing_pagination and tag == 'ul':
            self.__parsing_pagination = False
        elif self.__parsing_next_page_section and tag == 'a':
            self.__parsing_next_page_section = False


class MaterielNetParser(ProductParser):
    def __init__(self, url=None):
        super().__init__()
        self._product = Product()
        self._parsing_product = False
        self._parsing_product_meta = False
        self._parsing_title = False
        self.__parsing_product_availability = False
        self.__stock_web_id = None
        self._parsing_availability = False
        self.__parsing_price_category = False
        self.__parsing_price_objects = False
        self._parsing_price = False
        self._parsing_pagination = False
        self.__active_page_found = False
        self.__parsing_next_page = False
        self._pagination_parsed = False
        if url:
            self._base_url = parse_base_url(url)
        else:
            self._base_url = 'https://www.materiel.net'

    @property
    def _parsing_web_availability(self):
        return self.__parsing_product_availability and self.__stock_web_id

    def _close_availability_parsing(self):
        self._parsing_availability = False
        self.__stock_web_id = None
        self.__parsing_product_availability = False

    def _close_product_meta_parsing(self):
        self._parsing_product_meta = False

    def _close_title_parsing(self):
        self._parsing_title = False

    def _close_price_parsing(self):
        self.__parsing_price_category = False
        self.__parsing_price_objects = False
        self._parsing_price = False

    def _close_product_parsing(self):
        self._parsing_product = False
        self.products.append(self._product)
        self._product = Product()

    def _close_pagination_parsing(self):
        self._parsing_pagination = False
        self._pagination_parsed = True

    @staticmethod
    def parse_price(string):
        currency = None
        if '€' in string:
            currency = 'EUR'
        price = int(''.join([i for i in string if i.isdigit()]))
        return price, currency

    def handle_starttag(self, tag, attrs):
        if not self._parsing_product and tag == 'li':
            for name, value in attrs:
                if name == 'class' and 'ajax-product-item' in value.split(' '):
                    self._parsing_product = True

        if not self._parsing_product_meta and tag == 'div':
            for name, value in attrs:
                if name == 'class' and value == 'c-product__meta':
                    self._parsing_product_meta = True
        elif self._parsing_product_meta:
            if tag == 'a':
                for name, value in attrs:
                    if name == 'href':
                        self._product.url = f'{self._base_url}{value}'
            elif tag == 'h2':
                for name, value in attrs:
                    if name == 'class' and value == 'c-product__title':
                        self._parsing_title = True
        if tag == 'div':
            for name, value in attrs:
                if not self.__parsing_product_availability and name == 'class' and value == 'c-product__availability':
                    self.__parsing_product_availability = True
                elif self.__parsing_product_availability and name == 'data-stock-web':
                    self.__stock_web_id = value
        elif tag == 'span' and self._parsing_web_availability:
            for name, value in attrs:
                availability_class_name = f'o-availability__value--stock_{self.__stock_web_id}'
                if name == 'class' and availability_class_name in value.split(' '):
                    self._parsing_availability = True
        if not self.__parsing_price_objects and tag == 'div':
            for name, value in attrs:
                if not self.__parsing_price_category and name == 'class' and value == 'c-product__prices':
                    self.__parsing_price_category = True
                elif self.__parsing_price_category and name == 'class' and 'o-product__prices' in value.split(' '):
                    self.__parsing_price_objects = True
        elif self.__parsing_price_objects and tag == 'span':
            for name, value in attrs:
                if name == 'class' and value == 'o-product__price':
                    self._parsing_price = True
        if not self._pagination_parsed:
            if not self._parsing_pagination and tag == 'ul':
                for name, value in attrs:
                    if name == 'class' and value == 'pagination':
                        self._parsing_pagination = True
            elif self._parsing_pagination and tag == 'li':
                for name, value in attrs:
                    values = value.split(' ')
                    if not self.__active_page_found and name == 'class' and 'page-item' in values \
                            and 'active' in values:
                        self.__active_page_found = True
                    elif self.__active_page_found and name == 'class' and 'page-item' in values:
                        self.__parsing_next_page = True
            elif self.__parsing_next_page and tag == 'a':
                for name, value in attrs:
                    if name == 'href':
                        self.next_page = f'{self._base_url}{value}'
                        self.__parsing_next_page = False
                        self._pagination_parsed = True

    def handle_endtag(self, tag):
        if self._parsing_product_meta and tag == 'div':
            self._close_product_meta_parsing()
        elif self._parsing_product and tag == 'li':
            self._close_product_parsing()
        elif self._parsing_pagination and tag == 'ul':
            self._close_pagination_parsing()

    def handle_data(self, data):
        last_tag = self.get_starttag_text()
        if self._parsing_title and last_tag.startswith('<h2'):
            self._product.name = data
            self._close_title_parsing()
        elif self._parsing_availability and last_tag.startswith('<span'):
            self._product.available = data != 'Rupture'
            self._close_availability_parsing()
        elif self._parsing_price:
            if last_tag.startswith('<span'):
                self._product.price, self._product.price_currency = self.parse_price(data)
            elif last_tag.startswith('<sup>'):
                self._product.price += int(data) / 100
                self._close_price_parsing()


class AlternateParser(ProductParser):
    def __init__(self, url=None):
        super().__init__()
        self._product = Product()
        if url:
            self._base_url = parse_base_url(url)
        else:
            self._base_url = 'https://www.alternate.be'
        self._parsing_row = False
        self._parsing_name = False
        self._parsing_price = False

    def handle_starttag(self, tag, attrs):
        if not self._parsing_row and tag == 'div':
            for name, value in attrs:
                if name == 'class' and value == 'listRow':
                    self._parsing_row = True
        elif self._parsing_row:
            if tag == 'a':
                for name, value in attrs:
                    if name == 'href' and not self._product.url:
                        self._product.url = self.parse_url(value)
            elif tag == 'span':
                if not self._parsing_name:
                    for name, value in attrs:
                        if name == 'class':
                            if value == 'name':
                                self._parsing_name = True
                elif self._parsing_name:
                    for name, value in attrs:
                        if name == 'class' and value == 'additional':
                            self._parsing_name = False
                if not self._parsing_price:
                    for name, value in attrs:
                        if name == 'class' and 'price' in value.split(' '):
                            self._parsing_price = True
            elif tag == 'strong':
                for name, value in attrs:
                    if name == 'class' and 'stockStatus' in value.split(' '):
                        values = value.split(' ')
                        available = 'available_unsure' not in values and 'preorder' not in values
                        self._product.available = available

    def handle_data(self, data):
        if self._parsing_name:
            data = data.replace('grafische kaart', '').strip()
            if data:
                if not self._product.name:
                    self._product.name = data
                else:
                    self._product.name += f' {data}'
        elif self._parsing_price:
            price, currency = self.parse_price(data)
            if price and currency:
                self._product.price = price
                self._product.price_currency = currency
                self._parsing_price = False

    def handle_endtag(self, tag):
        if tag == 'span' and self._parsing_price:
            self._parsing_price = False
        elif tag == 'div' and self._parsing_row and self._product.ok():
            self._parsing_row = False
            self.products.append(self._product)
            self._product = Product()

    @staticmethod
    def parse_price(string):
        currency = None
        if '€' in string:
            currency = 'EUR'
        price = int(''.join([i for i in string if i.isdigit()]))
        return price, currency

    def parse_url(self, string):
        string = string.split('?')[0]  # remove query string
        return f'{self._base_url}{string}'


class MineShopParser:
    def __init__(self, url=None):
        self.products = []
        self._product = Product()

    def feed(self, webpage):
        tags = self._find_products(webpage)
        for tag in tags:
            # product has at least a name
            name = self._parse_name(tag)
            if not name:
                continue
            self._product.name = name
            # parse all other attributes
            price, currency = self._parse_price(tag)
            self._product.price = price
            self._product.price_currency = currency
            self._product.url = self._parse_url(tag)
            self._product.available = self._parse_availability(tag)
            # then add product to list
            self.products.append(self._product)
            self._product = Product()

    @staticmethod
    def _find_products(webpage):
        soup = BeautifulSoup(webpage, features='lxml')
        products = []
        tags = soup.find_all('ul')
        for tag in tags:
            if 'products' in tag.get('class', []):
                for child in tag.children:
                    products.append(child)
        return products

    @staticmethod
    def _parse_name(product):
        title = product.find('h2')
        if type(title) is Tag:
            return title.text

    @staticmethod
    def _parse_price(product):
        tag = product.find('bdi')
        if type(tag) is Tag:
            string = tag.text
            if '€' in string:
                currency = 'EUR'
                string = string.replace('€', '').strip()
            price = float(string)
            return price, currency

    @staticmethod
    def _parse_url(product):
        tag = product.find('a')
        if type(tag) is Tag and tag.get('href'):
            return tag['href']

    @staticmethod
    def _parse_availability(product):
        tag = product.find('p')
        if type(tag) is Tag:
            attributes = tag.get('class', [])
            if 'stock' in attributes:
                attributes.remove('stock')
                availability = attributes[0]
                return availability != 'out-of-stock'
        return True
