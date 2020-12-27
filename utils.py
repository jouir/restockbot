from math import floor
from urllib.parse import urlparse


def format_timedelta(value, time_format='{days} days, {hours2}:{minutes2}:{seconds2}'):
    """
    Taken from https://github.com/frnhr/django_timedeltatemplatefilter and
    https://github.com/frnhr/django_timedeltatemplatefilter
    """
    if hasattr(value, 'seconds'):
        seconds = value.seconds + value.days * 24 * 3600
    else:
        seconds = int(value)

    seconds_total = seconds

    minutes = int(floor(seconds / 60))
    minutes_total = minutes
    seconds -= minutes * 60

    hours = int(floor(minutes / 60))
    hours_total = hours
    minutes -= hours * 60

    days = int(floor(hours / 24))
    days_total = days
    hours -= days * 24

    years = int(floor(days / 365))
    years_total = years
    days -= years * 365

    return time_format.format(**{
        'seconds': seconds,
        'seconds2': str(seconds).zfill(2),
        'minutes': minutes,
        'minutes2': str(minutes).zfill(2),
        'hours': hours,
        'hours2': str(hours).zfill(2),
        'days': days,
        'years': years,
        'seconds_total': seconds_total,
        'minutes_total': minutes_total,
        'hours_total': hours_total,
        'days_total': days_total,
        'years_total': years_total,
    })


def parse_base_url(url, include_scheme=True):
    result = urlparse(url)
    base_url = f'{result.scheme}://{result.netloc}' if include_scheme else result.netloc.replace('www.', '')
    return base_url
