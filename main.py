#!/usr/bin/env python3
import argparse
import logging

from config import extract_shops, read_config
from crawlers import (AlternateCrawler, LDLCCrawler, MaterielNetCrawler,
                      TopAchatCrawler)
from db import create_tables, list_shops, upsert_products, upsert_shops
from notifiers import TwitterNotifier

logger = logging.getLogger(__name__)


def parse_arguments():
    parser = argparse.ArgumentParser()
    parser.add_argument('-v', '--verbose', dest='loglevel', action='store_const', const=logging.INFO,
                        help='print more output')
    parser.add_argument('-d', '--debug', dest='loglevel', action='store_const', const=logging.DEBUG,
                        default=logging.WARNING, help='print even more output')
    parser.add_argument('-o', '--logfile', help='logging file location')
    parser.add_argument('-c', '--config', default='config.json', help='configuration file location')
    parser.add_argument('-N', '--disable-notifications', dest='disable_notifications', action='store_true',
                        help='Do not send notifications')
    args = parser.parse_args()
    return args


def setup_logging(args):
    log_format = '%(asctime)s %(levelname)s: %(message)s' if args.logfile else '%(levelname)s: %(message)s'
    logging.basicConfig(format=log_format, level=args.loglevel, filename=args.logfile)


def main():
    args = parse_arguments()
    setup_logging(args)
    config = read_config(args.config)
    create_tables()

    shops = extract_shops(config['urls'])
    upsert_shops(shops.keys())

    if args.disable_notifications:
        notifier = None
    else:
        notifier = TwitterNotifier(consumer_key=config['twitter']['consumer_key'],
                                   consumer_secret=config['twitter']['consumer_secret'],
                                   access_token=config['twitter']['access_token'],
                                   access_token_secret=config['twitter']['access_token_secret'])

    for shop in list_shops():
        logger.debug(f'processing {shop}')
        urls = shops.get(shop.name)
        if not urls:
            logger.warning(f'cannot find urls for shop {shop} in the configuration file')
            continue
        if shop.name == 'topachat.com':
            crawler = TopAchatCrawler(shop=shop, urls=urls)
        elif shop.name == 'ldlc.com':
            crawler = LDLCCrawler(shop=shop, urls=urls)
        elif shop.name == 'materiel.net':
            crawler = MaterielNetCrawler(shop=shop, urls=urls)
        elif shop.name == 'alternate.be':
            crawler = AlternateCrawler(shop=shop, urls=urls)
        else:
            logger.warning(f'shop {shop} not supported')
            continue
        upsert_products(products=crawler.products, notifier=notifier)


if __name__ == '__main__':
    main()
