
# Promify

Promify is a command-line tool that converts datastreams into Prometheus metrics. It is a refactored and enhanced version of the [promify-goss](https://github.com/waydago/promify-goss) project, with the capability to handle a wider range of input sources. Promify is designed to be a pipe-only program, providing a flexible and efficient way to handle input sources. We have also added a formatter interface, allowing for a pluggable approach to adding new sources. Promify currently defaults to the `goss` format, but you can add any other "Formatter" as long as you satisfy the "Formatter" interface.

With Promify, you can easily build in your own input source to reformat into Prometheus metrics, which can then be scraped by Prometheus via Node Exporters textfile_collector and visualized in Grafana, making it an ideal tool to improve monitoring of your systems and applications. 

Promify requires the input source to be piped, ensuring a smooth data flow and easy integration with other tools. The `name` flag is mandatory when running Promify. This ensures that the output file is identifiable and can be easily located. The `path` flag is optional. If not specified, Promify will use Node Exporter's default `textfile_collector` directory "/var/lib/node_exporter/textfile_collector".

Try it out today and see how it can help you improve your monitoring and alerting workflows.

## Key Features

- **Pipe-Only Input**: Promify requires the input source to be piped, ensuring a smooth data flow and easy integration with other tools.
- **Required Name Flag**: The `name` flag is mandatory when running Promify. This ensures that the output file is identifiable and can be easily located.
- **Optional Directory Flag**: The `path` flag is optional. If not specified, Promify will use Node Exporter's default `textfile_collector` directory "/var/lib/node_exporter/textfile_collector".
- **Modular Input Formats**: One of the major enhancements in Promify is the adding the formatter interface. Promify defaults to the `goss` format, however, you may choose to add "debugvarz" etc as an input source allowing for a more flexible and modular approach to the input source.

## Basic Usage

Promified doesn't have many options.

```bash
Usage of ./promify:
  -format string
        Format of the input data (default "goss")
  -name string
        Name your .prom with the extension
  -path string
        Where to store the .prom file (default "/var/lib/node_exporter/textfile_collector")
```

## Example Usage

```bash
$ cat examples/demo.json | ./promify                                      
name is required
```

An unspecified `-path` will use the default textfile_collector path shipped by node_exporter.

```bash
$ cat examples/demo.json | ./promify -name t.prom -path /tmp
```
In the above example we are using the default goss format and the output file will be named t.prom and stored in /tmp.

If your node_exporter deployment has a custom textfile_collector you will need to specify that path or update your fork of the go code to make your path the default and rebuild the program.

## Input Sources

In our demo.yaml gossfile test we are expecting the file /srv/down not to exist and http://httpbun.org/get to return a 200 respose.

```bash
$ goss -g ./examples/demo.yaml validate -f tap
1..2
ok 1 - File: /srv/down: exists: matches expectation: [false]
ok 2 - HTTP: http://httpbun.org/get: status: matches expectation: [200]
```

Below is the data returned with the json outputter. At first glance we can already see json exposes more details about each test.

```bash
$ goss -g ./examples/demo.yaml validate -f json -o pretty
{
    "results": [
        {
            "duration": 52102,
            "err": null,
            "expected": [
                "false"
            ],
            "found": [
                "false"
            ],
            "human": "",
            "meta": null,
            "property": "exists",
            "resource-id": "/srv/down",
            "resource-type": "File",
            "result": 0,
            "skipped": false,
            "successful": true,
            "summary-line": "File: /srv/down: exists: matches expectation: [false]",
            "test-type": 0,
            "title": ""
        },
        {
            "duration": 523689683,
            "err": null,
            "expected": [
                "200"
            ],
            "found": [
                "200"
            ],
            "human": "",
            "meta": null,
            "property": "status",
            "resource-id": "http://httpbun.org/get",
            "resource-type": "HTTP",
            "result": 0,
            "skipped": false,
            "successful": true,
            "summary-line": "HTTP: http://httpbun.org/get: status: matches expectation: [200]",
            "test-type": 0,
            "title": ""
        }
    ],
    "summary": {
        "failed-count": 0,
        "summary-line": "Count: 2, Failed: 0, Duration: 0.524s",
        "test-count": 2,
        "total-duration": 523920650
    }
}
```

Now if we inspect the output of the goss formatter we can see that its written as prometheus metrics.

```bash
$ cat examples/demo.json | ./promify -name t.prom -path ./ ; cat t.prom       
goss_result_file{property="/srv/down",resource="exists",skipped="false"} 0
goss_result_file_duration{property="/srv/down",resource="exists",skipped="false"} 52102
goss_result_http{property="http://httpbun.org/get",resource="status",skipped="false"} 0
goss_result_http_duration{property="http://httpbun.org/get",resource="status",skipped="false"} 523689683
goss_results_summary{textfile="t.prom",name="tested"} 2
goss_results_summary{textfile="t.prom",name="failed"} 0
goss_results_summary{textfile="t.prom",name="duration"} 523920650
```
## Using the Taskfile

This project uses a `Taskfile.yaml` for task running. The `Taskfile.yaml` includes tasks for cleaning, linting, testing, building, and intalling for the application.

Here are some of the tasks you can run:

- `task clean`: Removes the built binary and any linked files.
- `task lint`: Runs the linter on the Go source code.
- `task test`: Runs the Go tests.
- `task build`: Builds the Go application.
- `task install`: Installs the application to `/usr/local/bin/` (requires access to sudo).

For example, to build the application, you would run `task build`.

Note: You need to have the [Task](https://taskfile.dev/#/installation) task runner installed to use these tasks.

Thank you for using Promify! We hope it's been helpful for your projects. If you have any feedback or ideas for how we can improve it, please let us know by opening an issue on our GitHub repository.

We also welcome contributions from the community if you're interested in helping out. We appreciate any help we can get to make Promify even better.

