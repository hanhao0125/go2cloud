import os
import timeit
from functools import wraps

import torch
import torchvision.models as models
from PIL import Image
from torchvision import transforms

resnet = models.resnet101(pretrained=True)
transform = transforms.Compose([  # [1]
    transforms.Resize(256),  # [2]
    transforms.CenterCrop(224),  # [3]
    transforms.ToTensor(),  # [4]
    transforms.Normalize(  # [5]
        mean=[0.485, 0.456, 0.406],  # [6]
        std=[0.229, 0.224, 0.225]  # [7]
    )])


def func_time(function):
    @wraps(function)
    def func_time(*args, **kwargs):
        t0 = timeit.default_timer()
        result = function(*args, **kwargs)
        t1 = timeit.default_timer()
        print("Total running time: %s s" % (str(t1 - t0)))
        return result
    return func_time


def label():
    import json
    with open("imagenet_class_index.json", "r") as f:
        class_idx = json.load(f)
    idx2label = [class_idx[str(k)][1].lower() for k in range(len(class_idx))]
    return idx2label


idx2label = label()


@func_time
def predict(path="dog.png"):
    img = Image.open(path)
    if img.mode == "L":
        print("error shape:", img.mode)
        return
    img_t = transform(img)
    batch_t = torch.unsqueeze(img_t, 0)
    with open('imagenet_classes.txt') as f:
        labels = [line.strip() for line in f.readlines()]

    resnet.eval()

    # Third, carry out model inference
    out = resnet(batch_t)

    # Forth, print the top 5 classes predicted by the model
    _, indices = torch.sort(out, descending=True)
    percentage = torch.nn.functional.softmax(out, dim=1)[0] * 100

    # with open("./predict.txt", 'a+') as f:
    #     f.write(path)
    #     f.write("\n")
    #     for i in indices[0][:5]:
    #         f.write(f"{labels[i]},{percentage[i].item()}")
    #         f.write('\n')
    print("|".join([idx2label[i] for i in indices[0][:5]]))
    return {
        "path": path,
        "tag": "|".join([idx2label[i] for i in indices[0][:5]])
    }


def list_all_images():
    # 接受：图片（如何接受图片？需要提供一个类似文件服务器的功能，给定 URL，返回一张图片）
    # 为什么要提供文件服务器？如果服务全在一个主机上，直接读取本地文件就行了，但是考虑服务器是分布式的，各个服务之间通过消息中间件沟通
    # 所以在接受到图片后，应该直接将图片信息（如所在服务器，文件路径）加入到消息队列，python 从消息队列里取数据，然后消费
    # 简单起见，先全放一台机器，消息直接就是文件路径，
    # 处理：预测，打标签
    # 返回：成功与否
    dir = "/Users/hanhao/Downloads/ILSVRC2012_img_test/"
    cnt = 0
    l = [dir + p for p in os.listdir(dir)]

    return l


if __name__ == "__main__":
    predict()
    # print(list_all_images()[0])
