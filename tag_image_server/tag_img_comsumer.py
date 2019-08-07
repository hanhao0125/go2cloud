import datetime
import json
import os
import timeit
from functools import wraps

import nsq
import torch
import torchvision.models as models
from PIL import Image as IM
from sqlalchemy import DATETIME, Column, Integer, String, create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker
from torchvision import transforms

Base = declarative_base()
engine = create_engine('mysql+pymysql://root:root@localhost:3306/cloud')
DBSession = sessionmaker(bind=engine)
session = DBSession()
with open("imagenet_class_index.json", "r") as f:
    class_idx = json.load(f)
idx2label = [class_idx[str(k)][1].lower() for k in range(len(class_idx))]
mounted_path = "/Users/hanhao/server"

resnet = models.resnet101(pretrained=True)
transform = transforms.Compose([  # [1]
    transforms.Resize(256),  # [2]
    transforms.CenterCrop(224),  # [3]
    transforms.ToTensor(),  # [4]
    transforms.Normalize(  # [5]
        mean=[0.485, 0.456, 0.406],  # [6]
        std=[0.229, 0.224, 0.225]  # [7]
    )])


def predict(path="dog.png"):
    img = IM.open(path)
    if img.mode == "L":
        print("error shape:", img.mode)
        return
    img_t = transform(img)
    batch_t = torch.unsqueeze(img_t, 0)

    resnet.eval()

    out = resnet(batch_t)

    _, indices = torch.sort(out, descending=True)
    percentage = torch.nn.functional.softmax(out, dim=1)[0] * 100

    print("|".join([idx2label[i] for i in indices[0][:5]]))
    return {
        "tag": "|".join([idx2label[i] for i in indices[0][:5]])
    }


class Node(Base):
    __tablename__ = 'filenode'
    id = Column(Integer, primary_key=True)
    tag = Column(String(100))
    full_path = Column(String(500))

    def __str__(self):
        return f"{self.id}\t {self.path}"


def handler(message):
    message = message.body.decode()
    iid = message

    n = session.query(Node).get(int(iid))
    if n == None:
        print("err,no such node,", iid)
        return True
    r = predict(mounted_path + n.full_path)
    if r == None:
        return True

    n.tag = r['tag'].split('|')[0]
    session.add(n)
    session.commit()
    session.close()
    return True


def consume():
    r = nsq.Reader(message_handler=handler, nsqd_tcp_addresses=['127.0.0.1:4150'],
                   topic='tag', channel='a', lookupd_poll_interval=15)

    nsq.run()  # tornado.ioloop.IOLoop.instance().start()


if __name__ == "__main__":
    consume()
