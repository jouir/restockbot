import logging
from datetime import datetime

from sqlalchemy import (Boolean, Column, DateTime, Float, ForeignKey, Integer,
                        String, create_engine, exc)
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import relationship, sessionmaker

logger = logging.getLogger(__name__)


Base = declarative_base()
engine = create_engine('sqlite:///restock.db')
Session = sessionmaker(bind=engine, autoflush=False)


class Shop(Base):
    __tablename__ = 'shop'
    id = Column(Integer, primary_key=True)
    name = Column(String, unique=True, nullable=False)

    def __repr__(self):
        return f'Shop<{self.name}>'

    def __ne__(self, shop):
        return self.name != shop.name


class Product(Base):
    __tablename__ = 'product'
    id = Column(Integer, primary_key=True)
    name = Column(String, nullable=False)
    url = Column(String, nullable=False, unique=True)
    price = Column(Float, nullable=False)
    price_currency = Column(String, nullable=False)
    available = Column(Boolean, nullable=False)
    updated_at = Column(DateTime)
    tweet_id = Column(Integer, unique=True)
    shop_id = Column(Integer, ForeignKey('shop.id'), nullable=False)
    shop = relationship('Shop', foreign_keys=[shop_id])

    def __repr__(self):
        return f'Product<{self.name}@{self.shop.name}>'

    def __ne__(self, product):
        return self.name != product.name or self.price != product.price or self.available != product.available \
               or self.url != product.url or self.shop != product.shop

    def ok(self):
        return self.name and self.url and self.price and self.price_currency and self.available is not None


def create_tables():
    Base.metadata.create_all(engine)
    logger.debug('tables created')


def list_shops():
    session = Session()
    shops = session.query(Shop).all()
    session.close()
    return shops


def upsert_shops(names):
    session = Session()
    try:
        for name in names:
            shop = Shop(name=name)
            query = session.query(Shop).filter(Shop.name == shop.name)
            shop_database = query.first()
            if not shop_database:
                logger.info(f'{shop} added')
                session.add(shop)
                session.commit()
                logger.debug('transaction committed')
    except exc.SQLAlchemyError:
        logger.exception('cannot commit transaction')
    finally:
        session.close()


def upsert_products(products, notifier=None):
    session = Session()
    try:
        for product in products:
            query = session.query(Product).filter(Product.name == product.name, Product.shop == product.shop)
            product_database = query.first()
            now = datetime.utcnow()
            tweet_id = None
            if not product_database:
                # product is new and available so we need to create an initial thread
                if notifier and product.available:
                    product.tweet_id = notifier.create_thread(product).id
                product.updated_at = now
                session.add(product)
                logger.info(f'{product} added')
            elif product != product_database:
                # notifications
                if notifier and product.available != product_database.available:
                    if product.available and not product_database.tweet_id:
                        # product is now available so we need to create an initial tweet (or thread)
                        tweet = notifier.create_thread(product)
                        if tweet:
                            tweet_id = tweet.id
                    elif not product.available and product_database.available and product_database.tweet_id:
                        # product is out of stock so we need to reply to previous tweet to close the thread
                        notifier.close_thread(tweet_id=product_database.tweet_id,
                                              duration=now-product_database.updated_at)
                query.update({Product.price: product.price, Product.price_currency: product.price_currency,
                              Product.available: product.available, Product.url: product.url,
                              Product.tweet_id: tweet_id, Product.updated_at: now})
                logger.info(f'{product} updated')
            session.commit()
            logger.debug('transaction committed')
    except exc.SQLAlchemyError:
        logger.exception('cannot commit transaction')
    finally:
        session.close()
