import nsq
from model import predict

import datetime
from sqlalchemy import Column, String, create_engine, Integer, DATETIME
from sqlalchemy.orm import sessionmaker
from sqlalchemy.ext.declarative import declarative_base

Base = declarative_base()


class Image(Base):
    __tablename__ = 'image'
    id = Column(Integer, primary_key=True)
    tag = Column(String(500))
    top5 = Column(String(500))
    path = Column(String(200))
    upath = Column(String(200))
    uploaddate = Column(DATETIME)

    def __str__(self):
        return f"{self.id}\t {self.path}"


engine = create_engine('mysql+pymysql://root:root@localhost:3306/cloud')
DBSession = sessionmaker(bind=engine)
session = DBSession()


def sql():
    s = session.query(Image).all()
    for i in s:
        print(i)


def handler(message):
    message = message.body.decode()
    iid, message = message.split("|")

    r = predict(message)
    if r == None:
        return True
    I = session.query(Image).get(int(iid))
    I.tag = r['tag'].split('|')[0]
    I.top5 = r['tag']
    session.commit()
    return True


def consume():
    r = nsq.Reader(message_handler=handler, nsqd_tcp_addresses=['127.0.0.1:4150'],
                   topic='tag', channel='a', lookupd_poll_interval=15)

    nsq.run()  # tornado.ioloop.IOLoop.instance().start()


if __name__ == "__main__":
    consume()
