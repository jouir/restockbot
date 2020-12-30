import logging

from parsers import (AlternateParser, LDLCParser, MaterielNetParser,
                     TopAchatParser)
from selenium import webdriver
from selenium.common.exceptions import TimeoutException
from selenium.webdriver.common.by import By
from selenium.webdriver.firefox.options import Options
from selenium.webdriver.support import expected_conditions
from selenium.webdriver.support.ui import WebDriverWait

logger = logging.getLogger(__name__)


class ProductCrawler(object):

    TIMEOUT = 3

    def __init__(self, shop):
        options = Options()
        options.headless = True
        self._driver = webdriver.Firefox(executable_path='/usr/local/bin/geckodriver', options=options)
        self._shop = shop
        self.products = []

    def __del__(self):
        self._driver.quit()

    def fetch(self, url, wait_for=None):
        self._driver.get(url)
        if wait_for:
            try:
                condition = expected_conditions.presence_of_element_located((By.CLASS_NAME, wait_for))
                WebDriverWait(self._driver, self.TIMEOUT).until(condition)
            except TimeoutException:
                logger.warning(f'timeout waiting for element "{wait_for}" at {url}')
        logger.info(f'url {url} fetched')
        webpage = self._driver.execute_script("return document.getElementsByTagName('html')[0].innerHTML")
        return webpage

    def add_shop(self, products):
        for product in products:
            product.shop = self._shop
        return products


class TopAchatCrawler(ProductCrawler):
    def __init__(self, shop, urls):
        super().__init__(shop)
        parser = TopAchatParser()
        for url in urls:
            webpage = self.fetch(url=url)
            parser.feed(webpage)
        self.products += self.add_shop(parser.products)


class LDLCCrawler(ProductCrawler):
    def __init__(self, shop, urls):
        super().__init__(shop)
        parser = LDLCParser()
        for url in urls:
            next_page = url
            previous_page = None
            while next_page != previous_page:
                webpage = self.fetch(url=next_page)
                parser.feed(webpage)
                previous_page = next_page
                next_page = parser.next_page
        self.products += self.add_shop(parser.products)


class MaterielNetCrawler(ProductCrawler):
    def __init__(self, shop, urls):
        super().__init__(shop)
        parser = MaterielNetParser()
        for url in urls:
            next_page = url
            previous_page = None
            while next_page != previous_page:
                webpage = self.fetch(url=next_page, wait_for='o-product__price')
                parser.feed(webpage)
                previous_page = next_page
                next_page = parser.next_page
        self.products += self.add_shop(parser.products)


class AlternateCrawler(ProductCrawler):
    def __init__(self, shop, urls):
        super().__init__(shop)
        parser = AlternateParser()
        for url in urls:
            webpage = self.fetch(url=url)
            parser.feed(webpage)
        self.products += self.add_shop(parser.products)


CRAWLERS = {
    'topachat.com': TopAchatCrawler,
    'ldlc.com': LDLCCrawler,
    'materiel.net': MaterielNetCrawler,
    'alternate.be': AlternateCrawler
}
