#!/usr/bin/env python3
import argparse
import logging
from concurrent import futures

from config import extract_shops, read_config
from crawlers import CRAWLERS
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
    parser.add_argument('-t', '--workers', type=int, help='number of workers for crawling')
    args = parser.parse_args()
    return args


def setup_logging(args):
    log_format = '%(asctime)s %(levelname)s: %(message)s' if args.logfile else '%(levelname)s: %(message)s'
    logging.basicConfig(format=log_format, level=args.loglevel, filename=args.logfile)


def crawl_shop(shop, urls):
    logger.debug(f'processing {shop}')
    crawler = CRAWLERS[shop.name](shop=shop, urls=urls)
    return crawler.products


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

    with futures.ThreadPoolExecutor(max_workers=args.workers) as executor:
        all_futures = []
        for shop in list_shops():
            urls = shops.get(shop.name)
            if not urls:
                logger.warning(f'cannot find urls for shop {shop} in the configuration file')
                continue
            all_futures.append(executor.submit(crawl_shop, shop, urls))
        for future in futures.as_completed(all_futures):
            products = future.result()
            upsert_products(products=products, notifier=notifier)


if __name__ == '__main__':
    main()
