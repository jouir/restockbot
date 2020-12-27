import json

from utils import parse_base_url


def read_config(filename):
    with open(filename, 'r') as fd:
        return json.load(fd)


def extract_shops(urls):
    """
    Parse shop name and return list of addresses for each shop
    Example: {"toto.com/first", "toto.com/second", "tata.com/first"}
          -> {"toto.com": ["toto.com/first", "toto.com/second"], "tata.com": ["tata.com/first"]}
    """
    result = {}
    for url in urls:
        base_url = parse_base_url(url, include_scheme=False)
        if base_url not in result:
            result[base_url] = [url]
        else:
            result[base_url].append(url)
    return result
