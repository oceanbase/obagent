# OBAgent

OBAgent is a monitor collection framework. OBAgent supplies pull and push mode data collection to meet different applications. By default, OBAgent supports these plugins: server data collection, OceanBase Database metrics collection, monitor data processing, and the HTTP service for Prometheus Protocol. To support data collection for other data sources, or customized data flow processes, you only need to develop plugins.

## Licencing

OBAgent is under [MulanPSL - 2.0](http://license.coscl.org.cn/MulanPSL2) license. You can freely copy and use the source code. When you modify or distribute the source code, please obey the MulanPSL - 2.0 license.

## Documentation

See [OBAgent Document](docs/about-obagent/what-is-obagent.md).

## How to get

### Dependencies

To build OBAgent, make sure that your Go version is 1.14 or above.

### From RPM package

OBAgent supplies RPM package. You can download it from the [Release page](https://mirrors.aliyun.com/oceanbase/community/stable/el/7/x86_64/) and install it by using this command:

```bash
rpm -ivh obagent-0.1-1.alios7.x86_64.rpm
```

### From source code

### Debug mode

```bash
make build // make build is debug mode by default
make build-debug
```

### Release mode

```bash
make build-release
```

## How to develop

You can develop plugins for OBAgent. For more information, see [Develop plugins for OBAgent](docs/develop-guide.md).

## Contributing

Contributions are warmly welcomed and greatly appreciated. Here are a few ways you can contribute:

- Raise us an [Issue](https://github.com/oceanbase/obagent/issues).

## Support

In case you have any problems when using OBAgent, welcome to reach out for help:

- [GitHub Issue](https://github.com/oceanbase/obagent/issues)
- [Official website](https://open.oceanbase.com/)
- Knowledge base [Coming soon]
