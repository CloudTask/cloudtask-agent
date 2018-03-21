# Cloudtask Agent APIs Manual

> `GET` - http://localhost:8600/cloudtask/v2/_ping

&nbsp;&nbsp;&nbsp;&nbsp; agent faq api, get node lcoal service config value and status info. 

``` json 
/*Response*/
HTTP 200 OK
{
    "app": "7969c05a25a3cdfd39645bba2d213173",
    "key": "509794cc-3539-4f58-a4d3-cc02a4f4848f",
    "node": {
        "type": 2,
        "hostname": "host.localdomain",
        "datacenter": "",
        "location": "myCluster",
        "os": "linux",
        "platform": "amd64",
        "ipaddr": "192.168.2.80",
        "apiaddr": "http://192.168.2.80:8600",
        "pid": 1,
        "singin": true,
        "timestamp": 1521633276,
        "alivestamp": 1521276530,
        "attach": "eyJqb2JtYXhjb3VudCI6MjU1fQo="
    },
    "systemconfig": {
        "version": "v.2.0.0",
        "pidfile": "./jobworker.pid",
        "retrystartup": true,
        "useserverconfig": true,
        "centerhost": "http://192.168.2.80:8985",
        "websitehost": "http://192.168.2.80:8091",
        "cluster": {
            "hosts": "192.168.2.80:2181,192.168.2.81:2181,192.168.2.82:2181",
            "root": "/cloudtask",
            "device": "",
            "runtime": "myCluster",
            "os": "",
            "platform": "",
            "pulse": "8s",
            "threshold": 1
        },
        "api": {
            "hosts": [
                ":8600"
            ],
            "enablecors": true
        },
        "cache": {
            "maxjobs": 255,
            "savedirectory": "./cache",
            "autoclean": true,
            "cleaninterval": "30m",
            "pullrecovery": "300s"
        },
        "logger": {
            "logfile": "./logs/jobworker.log",
            "loglevel": "info",
            "logsize": 20971520
        }
    },
    "cache": {
        "allocversion": 53,
        "jobstotal": 3
    }
}
```
> `GET` - http://localhost:8600/cloudtask/v2/jobs

&nbsp;&nbsp;&nbsp;&nbsp; get current node all jobs cache info.

``` json 
/*Response*/
HTTP 200 OK
{
    "content": "request successed.",
    "data": {
        "jobbase": [
            {
                "jobid": "399d4b159c65c9b34d2a3c41",
                "jobname": "ACCT.ShippingCost.job",
                "filename": "",
                "filecode": "d41d8cd98f00b204e9800998ecf8427e",
                "cmd": "./run.sh",
                "env": [],
                "timeout": 0,
                "version": 1,
                "schedule": [
                    {
                        "id": "1623e5f1a23",
                        "enabled": 1,
                        "turnmode": 2,
                        "interval": 2,
                        "startdate": "03/19/2018",
                        "enddate": "",
                        "starttime": "00:00",
                        "endtime": "23:59",
                        "selectat": "",
                        "monthlyof": {
                            "day": 1,
                            "week": ""
                        }
                    },
                    {
                        "id": "1623e5f1a23",
                        "enabled": 1,
                        "turnmode": 2,
                        "interval": 4,
                        "startdate": "03/19/2018",
                        "enddate": "",
                        "starttime": "00:00",
                        "endtime": "23:59",
                        "selectat": "",
                        "monthlyof": {
                            "day": 1,
                            "week": ""
                        }
                    }
                ]
            },
            {
                "jobid": "33bd7b52592f4f2c45262e3b",
                "jobname": "EDI.Portol",
                "filename": "cloudtask-2.2.1-GDEV-2018-03-21_15-54-34.tar.gz",
                "filecode": "f1c844efe815a8b0294d6c68479af9f4",
                "cmd": "ps aux",
                "env": [],
                "timeout": 0,
                "version": 3,
                "schedule": [
                    {
                        "id": "1623e6002ad",
                        "enabled": 1,
                        "turnmode": 2,
                        "interval": 2,
                        "startdate": "03/19/2018",
                        "enddate": "",
                        "starttime": "00:00",
                        "endtime": "23:59",
                        "selectat": "",
                        "monthlyof": {
                            "day": 1,
                            "week": ""
                        }
                    },
                    {
                        "id": "1623e6002ad",
                        "enabled": 1,
                        "turnmode": 2,
                        "interval": 4,
                        "startdate": "03/19/2018",
                        "enddate": "",
                        "starttime": "00:00",
                        "endtime": "23:59",
                        "selectat": "",
                        "monthlyof": {
                            "day": 1,
                            "week": ""
                        }
                    }
                ]
            },
            {
                "jobid": "72ec7bb9decf1e8ea92ad3da",
                "jobname": "MKPL-File-Update",
                "filename": "cloudtask-2.2.1-GDEV-2018-03-21_15-54-34.tar.gz",
                "filecode": "f1c844efe815a8b0294d6c68479af9f4",
                "cmd": "ps aux",
                "env": [],
                "timeout": 0,
                "version": 1,
                "schedule": [
                    {
                        "id": "4c83862942feefb4fff4e422",
                        "enabled": 1,
                        "turnmode": 2,
                        "interval": 2,
                        "startdate": "03/19/2018",
                        "enddate": "",
                        "starttime": "00:00",
                        "endtime": "23:59",
                        "selectat": "",
                        "monthlyof": {
                            "day": 1,
                            "week": ""
                        }
                    }
                ]
            }
        ]
    }
}
```

> `GET` - http://localhost:8600/cloudtask/v2/jobs/{jobid}

&nbsp;&nbsp;&nbsp;&nbsp; get current node single job cache info.

``` json 
/*Response*/
HTTP 200 OK
{
    "content": "request successed.",
    "data": {
        "jobbase": {
            "jobid": "399d4b159c65c9b34d2a3c41",
            "jobname": "ACCT.ShippingCost.job",
            "filename": "",
            "filecode": "d41d8cd98f00b204e9800998ecf8427e",
            "cmd": "./run.sh",
            "env": [],
            "timeout": 0,
            "version": 1,
            "schedule": [
                {
                    "id": "1623e5f1a23",
                    "enabled": 1,
                    "turnmode": 2,
                    "interval": 2,
                    "startdate": "03/19/2018",
                    "enddate": "",
                    "starttime": "00:00",
                    "endtime": "23:59",
                    "selectat": "",
                    "monthlyof": {
                        "day": 1,
                        "week": ""
                    }
                }
            ]
        }
    }
}
```

> `PUT` - http://localhost:8600/cloudtask/v2/jobs/action

&nbsp;&nbsp;&nbsp;&nbsp; action a job, operation is `start` | `stop`.

``` json
/*Request*/
{
    "runtime": "myCluster",
    "jobid": "8fee1ea957b7b6b49bd4e75f",
    "action": "start"
}

/*Response*/
HTTP 202 Accepted
{
    "content": "request accepted."
}
```
