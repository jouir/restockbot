#!/usr/bin/env python3

import json
from urllib.parse import urlparse

import tweepy


def main():
    with open('config.json', 'r') as fd:
        config = json.load(fd)

    if 'access_token' in config['twitter'] and 'access_token_secret' in config['twitter']:
        access_token = config['twitter']['access_token']
        access_token_secret = config['twitter']['access_token_secret']
    else:
        consumer_key = config['twitter']['consumer_key']
        consumer_secret = config['twitter']['consumer_secret']

        auth = tweepy.OAuthHandler(consumer_key, consumer_secret)

        try:
            redirect_url = auth.get_authorization_url()
            print(f'Please go to {redirect_url}')
        except tweepy.TweepError:
            print('Failed to get request token')

        token = urlparse(redirect_url).query.split('=')[1]

        verifier = input('Verifier:')
        auth.request_token = {'oauth_token': token, 'oauth_token_secret': verifier}

        try:
            auth.get_access_token(verifier)
        except tweepy.TweepError:
            print('Failed to get access token')

        access_token = auth.access_token
        access_token_secret = auth.access_token_secret

    print(f'access_token = {access_token}')
    print(f'access_token_secret = {access_token_secret}')


if __name__ == '__main__':
    main()
