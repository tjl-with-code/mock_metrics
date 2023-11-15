# fake-metrics

可以导入prometheus或者各类exporter的指标，并可以对导入指标的value和label进行修改，方便模拟和测试告警以及查询语句

## feature

### import metrics from prometheus,exporter or other endponit with prometheus metrics format

[![qLJYRK.png](https://s1.ax1x.com/2022/04/05/qLJYRK.png)](https://imgtu.com/i/qLJYRK)

支持两种方式导入metrics，一种是直接复制metrics的文本，另一种是使用url导入，url可以填写exporter或者prometheus或者任何具有prometheus metrics格式的endponit，但url需保证能够被fake-metrics部署的环境访问到。

### modify metrics value and label

[![qLJDII.png](https://s1.ax1x.com/2022/04/05/qLJDII.png)](https://imgtu.com/i/qLJDII)

导入metrics之后，可以方便的修改指标的值及其label

通过修改值可以方便的模拟告警的情况，来测试告警表达式的正确与否，同时，在开发环境也可以方便的mock数据，以供开发环境做展示等。

### preview metrics value like normal exporter

通过访问/metrics endponit即可预览导入的metrics，格式如同其他exporter一样

## deploy

可以直接使用k8s.yaml进行部署
```bash
kubectl apply -f k8s.yaml
```

## local test

可以直接使用docker进行部署，目前镜像已经推送到开发环境的harbor
```bash
docker run -d --name fake-metrics -p 8080:8080 harbor-dev.eecos.cn:1443/ecf/fake-metrics:0.2
```

然后打开浏览器，访问`http://localhost:8080`即可打开

## other

design & coding by ggq