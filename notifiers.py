import logging

import tweepy
from utils import format_timedelta

logger = logging.getLogger(__name__)


class TwitterNotifier(object):

    _hashtags_map = {
        'rtx 3060 ti': ['#nvidia', '#rtx3060ti'],
        'rtx 3070': ['#nvidia', '#rtx3070'],
        'rtx 3080': ['#nvidia', '#rtx3080'],
        'rtx 3090': ['#nvidia', '#rtx3090'],
        'rx 6800 xt': ['#amd', '#rx6800xt'],
        'rx 6800': ['#amd', '#rx6800'],
        'rx 5700 xt': ['#amd', '#rx5700xt'],
    }

    _currency_map = {
        'EUR': '€'
    }

    def __init__(self, consumer_key, consumer_secret, access_token, access_token_secret):
        auth = tweepy.OAuthHandler(consumer_key, consumer_secret)
        auth.set_access_token(access_token, access_token_secret)
        self._api = tweepy.API(auth)

    def create_thread(self, product):
        currency_sign = self._currency_map[product.price_currency]
        shop_name = product.shop.name
        price = f'{product.price}{currency_sign}'
        message = f'{shop_name}: {product.name} for {price} is available at {product.url}'
        hashtags = self._parse_hashtags(product)
        if hashtags:
            message += f' {hashtags}'
        return self._create_tweet(message=message)

    def close_thread(self, tweet_id, duration):
        thread = self._api.get_status(id=tweet_id)
        duration = format_timedelta(duration, '{hours_total}h{minutes2}m')
        message = f'''@{thread.user.screen_name} And it's over ({duration})'''
        return self._create_tweet(message=message, tweet_id=tweet_id)

    def _create_tweet(self, message, tweet_id=None):
        try:
            tweet = self._api.update_status(status=message, in_reply_to_status_id=tweet_id)
            logger.info(f'tweet {tweet.id} sent with message "{message}"')
            return tweet
        except tweepy.error.TweepError as err:
            logger.warning('cannot send tweet with message "{message}"')
            logger.warning(str(err))

    def _parse_hashtags(self, product):
        for patterns in self._hashtags_map:
            if all(elem in product.name.lower().split(' ') for elem in patterns.split(' ')):
                return ' '.join(self._hashtags_map[patterns])
