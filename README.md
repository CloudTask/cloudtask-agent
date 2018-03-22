# cloudtask-agent
The cloudtask platform task work node.


Join to the cloudtask runtime cluster, according to the runtime tasks distribution table information, 
responsibled for the final execution of the tasks.

### Documents 
* [APIs Manual](./APIs.md)
* [Configuration Introduction](./Configuration.md)

### Usage

> binary

``` bash
$  ./cloudtask-agent -f etc/config.yaml
```

> docker image
[![](https://images.microbadger.com/badges/image/cloudtask/cloudtask-agent:2.0.0.svg)](https://microbadger.com/images/cloudtask/cloudtask-agent:2.0.0 "Get your own image badge on microbadger.com")
[![](https://images.microbadger.com/badges/version/cloudtask/cloudtask-agent:2.0.0.svg)](https://microbadger.com/images/cloudtask/cloudtask-agent:2.0.0 "Get your own version badge on microbadger.com")
``` bash
$ docker run -d --net=host --restart=always \
  -v /opt/app/cloudtask-agent/etc/config.yaml:/opt/cloudtask/etc/config.yaml \
  -v /opt/app/cloudtask-agent/logs:/opt/cloudtask/logs \
  -v /opt/app/cloudtask-agent/cache:/opt/cloudtask/cache \
  -v /etc/localtime:/etc/localtime \
  --name=cloudtask-agnet \
  cloudtask/cloudtask-agent:2.0.0
```

## License
cloudtask source code is licensed under the [Apache Licence 2.0](http://www.apache.org/licenses/LICENSE-2.0.html). 

