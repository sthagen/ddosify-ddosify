<h1 align="center">
    <img src="https://raw.githubusercontent.com/ddosify/ddosify/master/assets/ddosify-logo-db.svg#gh-dark-mode-only" alt="Ddosify logo dark" width="336px" /><br />
    <img src="https://raw.githubusercontent.com/ddosify/ddosify/master/assets/ddosify-logo-wb.svg#gh-light-mode-only" alt="Ddosify logo light" width="336px" /><br />
    Ddosify - High-performance load testing tool
</h1>

<p align="center">
    <a href="https://github.com/ddosify/ddosify/releases" target="_blank"><img src="https://img.shields.io/github/v/release/ddosify/ddosify?style=for-the-badge&logo=github&color=orange" alt="ddosify latest version" /></a>&nbsp;
    <a href="https://github.com/ddosify/ddosify/actions/workflows/test.yml" target="_blank"><img src="https://img.shields.io/github/actions/workflow/status/ddosify/ddosify/test.yml?branch=master&style=for-the-badge&logo=github" alt="ddosify build result" /></a>&nbsp;
    <a href="https://pkg.go.dev/go.ddosify.com/ddosify" target="_blank"><img src="https://img.shields.io/github/go-mod/go-version/ddosify/ddosify?style=for-the-badge&logo=go" alt="golang version" /></a>&nbsp;
    <a href="https://app.codecov.io/gh/ddosify/ddosify" target="_blank"><img src="https://img.shields.io/codecov/c/github/ddosify/ddosify?style=for-the-badge&logo=codecov" alt="go coverage" /></a>&nbsp;
    <a href="https://goreportcard.com/report/github.com/ddosify/ddosify" target="_blank"><img src="https://goreportcard.com/badge/github.com/ddosify/ddosify?style=for-the-badge&logo=go" alt="go report" /></a>&nbsp;
    <a href="https://github.com/ddosify/ddosify/blob/master/LICENSE" target="_blank"><img src="https://img.shields.io/badge/LICENSE-AGPL--3.0-orange?style=for-the-badge&logo=none" alt="ddosify license" /></a>
    <a href="https://discord.gg/9KdnrSUZQg" target="_blank"><img src="https://img.shields.io/discord/898523141788287017?style=for-the-badge&logo=discord&label=DISCORD" alt="ddosify discord server" /></a>
    <a href="https://hub.docker.com/r/ddosify/ddosify" target="_blank"><img src="https://img.shields.io/docker/v/ddosify/ddosify?style=for-the-badge&logo=docker&label=docker&sort=semver" alt="ddosify docker image" /></a>
</p>


<p align="center">
<img src="https://raw.githubusercontent.com/ddosify/ddosify/master/assets/ddosify-quick-start.gif" alt="Ddosify - High-performance load testing tool quick start" />
</p>


## Features

✅ **[Scenario-Based](#config-file)** - Create your flow in a JSON file. Without a line of code!

✅ **[Different Load Types](#load-types)** - Test your system's limits across different load types.

✅ **[Parameterization](#parameterization-dynamic-variables)** - Use dynamic variables just like on Postman.

✅ **[Correlation](#correlation)** - Extract variables from earlier phases and pass them on to the following ones.

✅ **[Test Data](#test-data-set)** - Import test data from CSV and use it in the scenario.

✅ **[Assertion](#assertion)** - Verify that the response matches your expectations.

✅ **[Success Criteria](#success-criteria-pass--fail)** - Set the success criteria for your test.

✅ **[Cookies](#cookies)** - Pass cookies through steps and set initial cookies if you want.

✅ **Widely Used Protocols** - Currently supporting *HTTP, HTTPS, HTTP/2*. Other protocols are on the way.


## Installation

`ddosify` is available via [Docker](https://hub.docker.com/r/ddosify/ddosify), [Docker Extension](https://hub.docker.com/extensions/ddosify/ddosify-docker-extension), [Homebrew Tap](#homebrew-tap-macos-and-linux), and downloadable pre-compiled binaries from the [releases page](https://github.com/ddosify/ddosify/releases/latest) for macOS, Linux and Windows. For auto-completion, see: [Ddosify Completions](https://github.com/ddosify/ddosify/blob/master/completions/README.md).

### Docker

```bash
docker run -it --rm ddosify/ddosify
```

### Docker Extension

Run Ddosify open-source on Docker Desktop with Ddosify Docker extension. More: [https://hub.docker.com/extensions/ddosify/ddosify-docker-extension](https://hub.docker.com/extensions/ddosify/ddosify-docker-extension)

### Homebrew Tap (macOS and Linux)

```bash
brew install ddosify/tap/ddosify
```

### apk, deb, rpm, Arch Linux, FreeBSD packages

- For arm architectures change `ddosify_amd64` to `ddosify_arm64` or `ddosify_armv6`.
- Superuser privilege is required.

```bash
# For Redhat based (Fedora, CentOS, RHEL, etc.)
rpm -i https://github.com/ddosify/ddosify/releases/latest/download/ddosify_amd64.rpm

# For Debian based (Ubuntu, Linux Mint, etc.)
wget https://github.com/ddosify/ddosify/releases/latest/download/ddosify_amd64.deb
dpkg -i ddosify_amd64.deb

# For Alpine
wget https://github.com/ddosify/ddosify/releases/latest/download/ddosify_amd64.apk
apk add --allow-untrusted ddosify_amd64.apk

# For Arch Linux
git clone https://aur.archlinux.org/ddosify.git
cd ddosify
makepkg -sri

# For FreeBSD
pkg install ddosify
```

### Windows exe from the [releases page](https://github.com/ddosify/ddosify/releases/latest)

- Download *.zip file for your architecture. For example download ddosify version vx.x.x with amd64 architecture: ddosify_x.x.x.zip_windows_amd64
- Unzip `ddosify_x.x.x_windows_amd64.zip`
- Open Powershell or CMD (Command Prompt) and change directory to unzipped folder: `ddosify_x.x.x_windows_amd64`
- Run ddosify:

```bash
.\ddosify.exe -t http://target_site.com
```

### Using the convenience script (macOS and Linux)

- The script requires root or sudo privileges to move ddosify binary to `/usr/local/bin`.
- The script attempts to detect your operating system (macOS or Linux) and architecture (arm64, x86, amd64) to download the appropriate binary from the [releases page](https://github.com/ddosify/ddosify/releases/latest).
- By default, the script installs the latest version of `ddosify`.
- If you have problems, check [common issues](#common-issues)
- Required packages: `curl` and `sudo`

```bash
curl -sSfL https://raw.githubusercontent.com/ddosify/ddosify/master/scripts/install.sh | sh
```

### Go install from source (macOS, FreeBSD, Linux, Windows)

*Minimum supported Go version is 1.18*

```bash
go install -v go.ddosify.com/ddosify@latest
```

## Easy Start
This section aims to show you how to use Ddosify without deep dive into its details easily.

1. ### Simple load test

   	ddosify -t http://target_site.com

   The above command runs a load test with the default value that is 100 requests in 10 seconds.

2. ### Using some of the features

   	ddosify -t http://target_site.com -n 1000 -d 20 -m PUT -T 7 -P http://proxy_server.com:80

   Ddosify sends a total of *1000* *PUT* requests to *https://target_site.com* over proxy *http://proxy_server.com:80* in *20* seconds with a timeout of *7* seconds per request.

3. ### Usage for CI/CD pipelines (JSON output)

    	ddosify -t http://target_site.com -o stdout-json | jq .avg_duration

   Ddosify outputs the result in JSON format. Then `jq` (or any other command-line JSON processor) fetches the `avg_duration`. The rest depends on your CI/CD flow logic.

4. ### Scenario based load test

   	ddosify -config config_examples/config.json
   Ddosify first sends *HTTP/2 POST* request to *https://test_site1.com/endpoint_1* using basic auth credentials *test_user:12345* over proxy *http://proxy_host.com:proxy_port*  and with a timeout of *3* seconds. Once the response is received, HTTPS GET request will be sent to *https://test_site1.com/endpoint_2* along with the payload included in *config_examples/payload.txt* file with a timeout of 2 seconds. This flow will be repeated *20* times in *5* seconds and response will be written to *stdout*.

5. ### Load test with Dynamic Variables (Parameterization)

    	ddosify -t http://target_site.com/{{_randomInt}} -d 10 -n 100 -h 'User-Agent: {{_randomUserAgent}}' -b '{"city": "{{_randomCity}}"}'
   Ddosify sends a total of *100* *GET* requests to *https://target_site.com/{{_randomInt}}* in *10* seconds. `{{_randomInt}}` path generates random integers between 1 and 1000 in every request. Dynamic variables can be used in *URL*, *headers*, *payload (body)* and *basic authentication*. In this example, Ddosify generates a random user agent in the header and a random city in the body. The full list of the dynamic variables can be found in the [docs](https://docs.ddosify.com/extra/dynamic-variables-parameterization).

6. ### Correlation (Captured Variables)

    	ddosify -config ddosify_config_correlation.json
   Ddosify allows you to specify variables at the global level and use them throughout the scenario, as well as extract variables from previous steps and inject them to the next steps in each iteration individually. You can inject those variables in requests *url*, *headers* and *payload(body)*. The example config can be found in [correlation-config-example](#Correlation).

7. ### Test Data

    	ddosify -config ddosify_data_csv.json
   Ddosify allows you to load test data from a file, tag specific columns for later use. You can inject those variables in requests *url*, *headers* and *payload(body)*. The example config can be found in [test-data-example](#test-data-set).
## Details

You can configure your load test by the CLI options or a config file. Config file supports more features than the CLI. For example, you can't create a scenario-based load test with CLI options.
### CLI Flags

```bash
ddosify [FLAG]
```

| Flag | Description                  | Type     | Default | Required?  |
| ------ | -------------------------------------------------------- | ------   | ------- | ---------  |
| `-t`   | Target website URL. Example: https://ddosify.com         | `string` | - | Yes        |
| `-n`   | Total iteration count                                      | `int`    | `100`   | No         |
| `-d`   | Test duration in seconds.                                | `int`    | `10`    | No         |
| `-m`   | Request method. Available methods for HTTP(s) are *GET, POST, PUT, DELETE, HEAD, PATCH, OPTIONS* | `string`    | `GET`    | No  |
| `-b`   | The payload of the network packet. AKA body for the HTTP.  | `string`    | -    | No         |
| `-a`   | Basic authentication. Usage: `-a username:password`        | `string`    | -    | No         |
| `-h`   | Headers of the request. You can provide multiple headers with multiple `-h` flag. Usage: `-h 'Accept: text/html'`  | `string`| -    | No         |
| `-T`   | Timeout of the request in seconds.                       | `int`    | `5`    | No         |
| `-P`   | Proxy address as host:port. `-P 'http://user:pass@proxy_host.com:port'` | `string`    | -    | No |
| `-o`   | Test result output destination. Supported outputs are [*stdout, stdout-json*] Other output types will be added. | `string`    | `stdout`    | No |
| `-l`   | [Type](#load-types) of the load test. Ddosify supports 3 load types. | `string`    | `linear`    | No |
| <span style="white-space: nowrap;">`--config`</span>    | [Config File](#config-file) of the load test. | `string`    | -    | No |
| <span style="white-space: nowrap;">`--version`</span>    | Prints version, git commit, built date (utc), go information and quit | -    | -    | No |
| <span style="white-space: nowrap;">`--cert_path`</span>    | A path to a certificate file (usually called 'cert.pem') | -    | -    | No |
| <span style="white-space: nowrap;">`--cert_key_path`</span>    | A path to a certificate key file (usually called 'key.pem') | -    | -    | No |
| <span style="white-space: nowrap;">`--debug`</span>    | Iterates the scenario once and prints curl-like verbose result. Note that this flag overrides json config.  |  `bool`     |  `false`     | No |

### Load Types

#### Linear

```bash
ddosify -t http://target_site.com -l linear
```

Result:

![linear load](https://raw.githubusercontent.com/ddosify/ddosify/master/assets/linear.gif)

*Note:* If the iteration count is too low for the given duration, the test might be finished earlier than you expect.

#### Incremental

```bash
ddosify -t http://target_site.com -l incremental
```

Result:

![incremental load](https://raw.githubusercontent.com/ddosify/ddosify/master/assets/incremental.gif)


#### Waved

```bash
ddosify -t http://target_site.com -l waved
```

Result:

![waved load](https://raw.githubusercontent.com/ddosify/ddosify/master/assets/waved.gif)


### Config File

Config file lets you use all capabilities of Ddosify.

The features you can use by config file;

- Scenario creation
- Environment variables
- Correlation
- Assertions
- Cookies
- Custom load type creation
- Payload from a file
- Multipart/form-data payload
- Extra connection configuration
- HTTP2 support


Usage;

    ddosify -config <json_config_path>


There is an example config file at [config_examples/config.json](/config_examples/config.json). This file contains all of the parameters you can use. Details of each parameter;

- `iteration_count` *optional*

  This is the equivalent of the `-n` flag. The difference is that if you have multiple steps in your scenario, this value represents the iteration count of the steps.

- `load_type` *optional*

  This is the equivalent of the `-l` flag.

- `duration` *optional*

  This is the equivalent of the `-d` flag.

- `manual_load` *optional*

  If you are looking for creating your own custom load type, you can use this feature. The example below says that Ddosify will run the scenario 5 times, 10 times, and 20 times, respectively along with the provided durations. `iteration_count` and `duration` will be auto-filled by Ddosify according to `manual_load` configuration. In this example, `iteration_count` will be 35 and the `duration` will be 18 seconds.
  Also `manual_load` overrides `load_type` if you provide both of them. As a result, you don't need to provide these 3 parameters when using `manual_load`.
    ```json
    "manual_load": [
        {"duration": 5, "count": 5},
        {"duration": 6, "count": 10},
        {"duration": 7, "count": 20}
    ]
    ```

- `proxy` *optional*

  This is the equivalent of the `-P` flag.

- `output` *optional*

  This is the equivalent of the `-o` flag.
- `engine_mode` *optional*
  Can be one of `distinct-user`, `repeated-user`, or default mode `ddosify`.
    - `distinct-user` mode simulates a new user for every iteration.
    - `repeated-user` mode can use pre-used user in subsequent iterations.
    - `ddosify` mode is default mode of the engine. In this mode engine runs in its max capacity, and does not show user simulation behaviour.

- `env` *optional*
  Scenario-scoped global variables. Note that dynamic variables changes every iteration.
    ```json
    "env": {
            "COMPANY_NAME" :"Ddosify",
            "randomCountry" : "{{_randomCountry}}"
    }
    ``` 

- `data` *optional*
  Config for loading test data from a CSV file.
  [CSV data](https://github.com/ddosify/ddosify/tree/master/config/config_testdata/test.csv) used in below config.
    ```json
    "data":{
        "info": {
            "path" : "config/config_testdata/test.csv",
            "delimiter": ";",
            "vars": {
                    "0":{"tag":"name"},
                    "1":{"tag":"city"},
                    "2":{"tag":"team"},
                    "3":{"tag":"payload", "type":"json"},
                    "4":{"tag":"age", "type":"int"}
                    },
            "allow_quota" : true,
            "order": "sequential",
            "skip_first_line" : true,
            "skip_empty_line" : true
        }
    }
    ```
  | Field | Description                  | Type     | Default | Required?  |
  | ------ | -------------------------------------------------------- | ------   | ------- | ---------  |
  | `path`   | Local path or remote url for your CSV file         | `string` | - | Yes        |
  | `delimiter`   | Delimiter for reading CSV                                      | `string`    | `,`   | No         |
  | `vars`   | Tag columns using column index as key, use `type` field if you want to cast a column to a specific type, default is `string`, can be one of the following: `json`, `int`, `float`,`bool`.                          | `map`    | -    | Yes         |
  | `allow_quota`   | If set to true, a quote may appear in an unquoted field and a non-doubled quote may appear in a quoted field | `bool`    | `false`    | No  |
  | `order`   | Order of reading records from CSV. Can be `random` or `sequential`                                | `string`    | `random`    | No         |
  | `skip_first_line`   | Skips first line while reading records from CSV.                                | `bool`    | `false`    | No         |
  | `skip_empty_line`   | Skips empty lines while reading records from CSV.                                | `bool`    | `true`    | No         |

- `success_criterias` *optional*
  Config for pass fail logic for the test. *abort* and *delay* fields can be used to adjust the abort behaviour in case of failure. If abort is true for a rule and rules fails at certain point, engine will decide to abort test immediately if delay is 0 or not given. If delay is given, it will wait for delay seconds and reassert the rule.

  **Example:** Check *90th percentile* and *fail_count*;
    ```json
    { 
     "duration": 10,
     <other_global_configurations>,
     "success_criterias": [
      {
        "rule" : "p90(iteration_duration) < 220", 
        "abort" : false
      },
      {
        "rule" : "fail_count_perc < 0.1", 
        "abort" : true,
        "delay" : 1
      },
      {
        "rule" : "fail_count < 100", 
        "abort" : true,
        "delay" : 0
      }
    ],
     "steps": [....]
    }
    ``` 

- `steps` *mandatory*

  This parameter lets you create your scenario. Ddosify runs the provided steps, respectively. For the given example file step id: 2 will be executed immediately after the response of step id: 1 is received. The order of the execution is the same as the order of the steps in the config file.

  **Details of each parameter for a step;**
    - `id` *mandatory*

      Each step must have a unique integer id.

    - `url` *mandatory*

      This is the equivalent of the `-t` flag.

    - `name` *optional* <a name="#step-name"></a>

      Name of the step.

    - `method` *optional*

      This is the equivalent of the `-m` flag.

    - `headers` *optional*

      List of headers with key:value format.

    - `payload` *optional*

      Body or payload. This is the equivalent of the `-b` flag.

      *Note:* If you want to use `x-www-form-urlencoded`, set Content-Type header to `application/x-www-form-urlencoded`.

      **Example:** send `x-www-form-urlencoded` data;
        ```json
        {
            "headers": {
                "Content-Type": "application/x-www-form-urlencoded"
            },
            "payload": "key1=value1&key2=value2"
        }
        ```

    - `payload_file` *optional*

      If you need a long payload, we suggest using this parameter instead of `payload`.

    - `payload_multipart` *optional* <a name="#payload_multipart"></a>

      Use this for `multipart/form-data` Content-Type.

      Accepts list of `form-field` objects, structured as below;
        ```json
        {
            "name": [field-name],
            "value": [field-value|file-path|url],
            "type": <text|file>,    // Default "text"
            "src": <local|remote>   // Default "local"
        }
        ```

      **Example:** Sending form name-value pairs;
        ```json
        "payload_multipart": [
            {
                "name": "[field-name]",
                "value": "[field-value]"
            }
        ]
        ```

      **Example:** Sending form name-value pairs and a local file;
        ```json
        "payload_multipart": [
            {
                "name": "[field-name]",
                "value": "[field-value]",
            },
            {
                "name": "[field-name]",
                "value": "./test.png",
                "type": "file"
            }
        ]
        ```

      **Example:** Sending form name-value pairs and a local file and a remote file;
        ```json
        "payload_multipart": [
            {
                "name": "[field-name]",
                "value": "[field-value]",
            },
            {
                "name": "[field-name]",
                "value": "./test.png",
                "type": "file"
            },
            {
                "name": "[field-name]",
                "value": "http://test.com/test.png",
                "type": "file",
                "src": "remote"
            }
        ]
        ```

      *Note:* Ddosify adds `Content-Type: multipart/form-data; boundary=[generated-boundary-value]` header to the request when using `payload_multipart`.

    - `timeout` *optional*

      This is the equivalent of the `-T` flag.

    - `capture_env` *optional*

      Config for extraction of variables to use them in next steps.
      **Example:** Capture *NUM* variable from steps response body;
        ```json
        "steps": [
            {
                "id": 1,
                "url": "http://target.com/endpoint1",
                "capture_env": {
                     "NUM" :{"from":"body","json_path":"num"},
                }
            },
        ]
        ``` 
    - `assertion` *optional*

      The response from this step will be subject to the assertion rules. If one of the provided rules fails, step is considered as failure.
      **Example:** Check *status code* and *content-length* header values;
       ```json
       "steps": [
           {
               "id": 1,
               "url": "http://target.com/endpoint1",
               "assertion": [
                   "equals(status_code,200)",
                   "in(headers.content-length,[2000,3000])"
               ]
           },
       ]
       ``` 

    - `sleep` *optional* <a name="#sleep"></a>

      Sleep duration(ms) before executing the next step. Can be an exact duration or a range.

      **Example:** Sleep 1000ms after step-1;
        ```json
        "steps": [
            {
                "id": 1,
                "url": "http://target.com/endpoint1",
                "sleep": "1000"
            },
            {
                "id": 2,
                "url": "http://target.com/endpoint2",
            }
        ]
        ```

      **Example:** Sleep between 300ms-500ms after step-1;
        ```json
        "steps": [
            {
                "id": 1,
                "url": "http://target.com/endpoint1",
                "sleep": "300-500"
            },
            {
                "id": 2,
                "url": "http://target.com/endpoint2",
            }
        ]
        ```

    - `auth` *optional*

      Basic authentication.
        ```json
        "auth": {
            "username": "test_user",
            "password": "12345"
        }
        ```
    - `others` *optional*

      This parameter accepts dynamic *key: value* pairs to configure connection details of the protocol in use.

        ```json
        "others": {
            "disable-compression": false,    // Default true
            "h2": true,                      // Enables HTTP/2. Default false.
            "disable-redirect": true         // Default false
        }
        ```

## Parameterization (Dynamic Variables)

Just like the Postman, Ddosify supports parameterization (dynamic variables) on *URL*, *headers*, *payload (body)* and *basic authentication*. Actually, we support all the random methods Postman supports. If you use `{{$randomVariable}}` on Postman you can use it as `{{_randomVariable}}` on Ddosify. Just change `$` to `_` and you will be fine. To simulate a realistic load test on your system, Ddosify can send every request with dynamic variables.

The full list of dynamic variables can be found in the [Ddosify Docs](https://docs.ddosify.com/extra/dynamic-variables-parameterization).

### Parameterization on URL

Ddosify sends *100* GET requests in *10* seconds with random string `key` parameter. This approach can be also used in cache bypass.

```bash
ddosify -t http://target_site.com/?key={{_randomString}} -d 10 -n 100
```

### Parameterization on Headers

Ddosify sends *100* GET requests in *10* seconds with random `Transaction-Type` and `Country` headers.

```bash
ddosify -t http://target_site.com -d 10 -n 100 -h 'Transaction-Type: {{_randomTransactionType}}' -h 'Country: {{_randomCountry}}'
```

### Parameterization on Payload (Body)

Ddosify sends *100* GET requests in *10* seconds with random `latitude` and `longitude` values in body.

```bash
ddosify -t http://target_site.com -d 10 -n 100 -b '{"latitude": "{{_randomLatitude}}", "longitude": "{{_randomLongitude}}"}'
```

### Parameterization on Basic Authentication

Ddosify sends *100* GET requests in *10* seconds with random `username` and `password` with basic authentication.

```bash
ddosify -t http://target_site.com -d 10 -n 100 -a '{{_randomUserName}}:{{_randomPassword}}'
```

### Parameterization on Config File

Dynamic variables can be used on config file as well. Ddosify sends *100* GET requests in *10* seconds with random string `key` parameter in URL and random `User-Key` header.

```bash
ddosify -config ddosify_config_dynamic.json
```

```json
{
    "iteration_count": 100,
    "load_type": "linear",
    "duration": 10,
    "steps": [
        {
            "id": 1,
            "url": "https://test_site1.com/?key={{_randomString}}",
            "method": "POST",
            "headers": {
                "User-Key": "{{_randomInt}}"
            }
        }
    ]
}
```

### Operating System Environment Variables

In addition, you can also use operating system environment variables. To access these variables, simply add the `$` prefix followed by the variable name wrapped in double curly braces. The syntax for this is `{{$OS_ENV_VARIABLE}}` within the **config file**. For instance, to use the `USER` environment variable from your operating system, simply input `{{$USER}}`. You can use operating system environment variables in `URL`, `Headers`, `Body (Payload)`, and `Basic Authentication`. Here is an example of using operating system environment variables in the config file. `TARGET_SITE` operating system environment variable is used in `URL` and `USER` environment variable is used in `Headers`.

```bash
export TARGET_SITE="https://test_site1.com"
ddosify -config ddosify_config_os_env.json
```

```json
{
    "iteration_count": 100,
    "load_type": "linear",
    "duration": 10,
    "steps": [
        {
            "id": 1,
            "url": "{{$TARGET_SITE}}",
            "method": "POST",
            "headers": {
                "os-env-user": "{{$USER}}"
            }
        }
    ]
}
```

## Assertion

At default, Ddosify marks the step result as successful if it sends the request and receives the response without any network error happening. Status code or body type (or content) does not have any effect on success/failure criteria. But this may not be a good test result for your use case and you may want to create your success/fail logic. That's where you can use Assertions.

Ddosify supports assertion on `status code`, `response body`, `response size`, `response time`, `headers`, and `variables`. You can use the `assertion` parameter on the config file to check if the response matches the given condition per step. If the condition is not met, Ddosify will fail the step. Check the [example config](https://github.com/ddosify/ddosify/blob/master/config_examples/config.json) to see how it looks.

As shown in the related table the first 5 keywords store different data related to the response. The last keyword `variables` stores the current state of environment variables for the Step. You can use [Functions](#functions) or [Operators](#operators) to build conditional expressions based on these keywords.

You can write multiple assertions for a step. If any assertion fails, the step is marked as Failed.

If Ddosify can't receive the response for a request, that step is marked as Failed without processing the assertions. You will see a **Server Error** as a failure reason on the test result instead of an **Assertion Error**.

### Keywords

| Keyword | Description                  | Usage | 
| ------ | -------------------------------------------------------- | ------ |
| `status_code`   | Status code      | - |
| `body`   | Response body      | - |
| `response_size`   | Response size in bytes    | -  |
| `response_time`   | Response time in ms     | -   |     
| `headers`   | Response headers         | headers.header-key    |      
| `variables`   | Global and captured variables                          | variables.VarName    |  

### Functions

| Function | Parameters|   Description  |             
| ------ | -------------------------------------------------------- | ------ |  
| `less_than`   | ( param `int`, limit `int` )   | checks if param is less than limit |
| `greater_than`   | ( param `int`, limit `int` )   | checks if param is greater than limit |
| `exists`   | ( param `any` ) | checks if variable exists |
| `equals`   | ( param1 `any`, param2 `any` ) | checks if given parameters are equal |
| `equals_on_file`   |    ( param `any`, file_path `string` )   | reads from given file path and checks if it equals to given parameter |
| `in`   | ( param `any`, array_param `array` ) | checks if expression is in given array |
| `contains`   | ( param1 `any`, param2 `any` ) | makes substring with param1 inside param2
| `not`   | ( param `bool` ) | returns converse of given param |
| `range`   | ( param `int`, low `int`,high `int` ) | returns param is in range of [low,high): low is included, high is not included. |
| `json_path`   | ( json_path `string`) | extracts from response body using given json path |
| `xml_path`   | ( xpath `string` ) | extracts from response body using given xml path |
| `regexp` | ( param `any`, regexp `string`, matchNo `int` ) | extracts from given value in the first parameter using given regular expression |

### Operators
| Operator | Description  |                
| ------- | ---------  |
| `==`   | equals      | 
| `!=`   | not equals |     
| `>`   | greater than    |                            
| `<`   | less than  |  
| `!`   | not|
| `&&`   | and|
| `\|\|`   | or |


### Assertion Examples

| Expression | Description   |               
| ------ | -------------------------------------------------------- |
| `less_than(status_code,201)`   | checks if status code is less than 201   |
| `equals(status_code,200)`   | checks if status code equals to 200      |
| `status_code == 200`   | same as preceding one  |
| `not(status_code == 500)`   | checks if status code not equals to 500   |
| `status_code != 500`   | same as preceding one|
| `equals(json_path(\"employees.0.name\"),\"Name\")`   | checks if json extracted value is equal to "Name"|
| `equals(xpath(\"//item/title\"),\"ABC\")`   | checks if xml extracted value is equal to "ABC" |
| `equals(variables.x,100)`   | checks if `x` variable coming from global or captured variables is equal to 100|
| `equals(variables.x,variables.y)`   | checks if variables `x` and `y` are equal to each other |
| `equals_on_file(body,\"file.json\")`   | reads from file.json and compares response body with read file |
| `exists(headers.Content-Type)`   | checks if content-type header exists in response headers|
| `contains(body,\"xyz\")`   | checks if body contains "xyz" in it|
| `range(headers.content-length,100,300)`   | checks if content-length header is in range [100,300) | 
| `in(status_code,[200,201])`   | checks if status code equal to 200 or 201     |
| `(status_code == 200) \|\| (status_code == 201)`   | same as preceding one |
| `regexp(body,\"[a-z]+_[0-9]+\",0) == \"messi_10\"`   | checks if matched result from regex is equal to "messi_10" |

## Success Criteria (Pass / Fail)

Ddosify supports success criteria, allowing users to verify the success of their load tests based on response times and failure counts of iterations. With this feature, users can assert the percentile of response times and the failure counts of all the iterations in a test.

Users can specify the required percentile of response times and failure counts in the configuration file, and the engine will compare the actual response times and failure counts to these values throughout the test continuously. According to users configuration test can be aborted or continue running until end. Check the [example config](https://github.com/ddosify/ddosify/blob/master/config_examples/config.json) to see how it looks the `success_criterias` keyword.

Note that the functions and operators mentioned in the [Step Assertion](#assertion) section can also be utilized for the Success Criteria keywords listed below.

You can see an success criteria example in [EXAMPLES](https://github.com/ddosify/ddosify/blob/master/engine_docs/EXAMPLES.md#example-2-success-criteria) file.

## Difference Between Success Criteria and Step Assertions

Unlike assertions focused on individual steps, which determine the success or failure of a step according to its response, Success Criteria create an abort/continue logic for the entire test, which is based on the accumulated data from all iterations. 

### Keywords


| Keyword              | Description                          | Usage                                                             |
| ---------------------- | -------------------------------------- | ------------------------------------------------------------------- |
| `fail_count`         | Failure count of iterations          | Used for aborting when test exceeds certain fail_count            |
| `iteration_duration` | Response times of iterations in ms   | Used for percentile functions                                     |
| `fail_count_perc`    | Fail count percentage, in range [0,1] | Used for aborting when test exceeds certain fail count percentage |

### Functions

| Function | Parameters           | Description                                     |
| ---------- | ---------------------- | ------------------------------------------------- |
| `p99`    | ( arr `int array` ) | 99th percentile, use as `p99(iteration_duration)` |
| `p98`    | ( arr `int array` ) | 98th percentile, use as `p98(iteration_duration)` |
| `p95`    | ( arr `int array`)  | 95th percentile, use as `p95(iteration_duration)` |
| `p90`    | ( arr `int array`)  | 90th percentile, use as `p90(iteration_duration)` |
| `p80`    | ( arr `int array`)  | 80th percentile, use as `p80(iteration_duration)` |
| `min`    | ( arr `int array`)  | returns minimum element                         |
| `max`    | ( arr `int array`)  | returns maximum element                         |
| `avg`    | ( arr `int array`)  | calculates and returns average                  |

### Examples
| Expression                        | Description                                               |
| ----------------------------------- | ----------------------------------------------------------- |
| `p95(iteration_duration) < 100`   | 95th percentile should be less than 100 ms                |
| `less_than(fail_count,120)`       | Total fail count should be less than 120 |
| `less_than(fail_count_perc,0.05)` | Fail count percentage should be less than 5%              |

## Correlation
Ddosify enables you to capture variables from steps using **json_path**, **xpath**, or **regular expressions**. Later, in the subsequent steps, you can inject both the captured variables and the scenario-scoped global variables.

> **:warning: Points to keep in mind**
> - You must specify **'header_key'** when capturing from header.
> - For json_path syntax, please take a look at [gjson syntax](https://github.com/tidwall/gjson/blob/master/SYNTAX.md) doc.
> - Regular expression are expected in  **'Golang'** style regex. For converting your existing regular expressions, you can use [regex101](https://regex101.com/).
> - You can extract values from **headers**, **body**, and **cookies**.

You can use **debug** parameter to validate your config.

```bash
ddosify -config ddosify_config_correlation.json -debug
```

### Capture With json_path
```json
{
    "steps": [
        {
            "capture_env": {
                "NUM" :{"from":"body","json_path":"num"},
                "NAME" :{"from":"body","json_path":"name"},
                "SQUAD" :{"from":"body","json_path":"squad"},
                "PLAYERS" :{"from":"body","json_path":"squad.players"},
                "MESSI" : {"from":"body","json_path":"squad.players.0"},             
            }         
        }
    ]
}
```

### Capture With XPath
```json
{
    "steps": [
        {
            "capture_env": {
                "TITLE" :{"from":"body","xpath":"//item/title"},             
            }         
        }
    ]
}
```

### Capture With Regular Expressions
```json
{
    "steps": [
        {
            "capture_env": {
               "CONTENT_TYPE": {"from":"header", "header_key":"Content-Type" ,"regexp":{"exp":"application\/(\\w)+","matchNo":0}} ,
               "REGEX_MATCH_ENV": {"from":"body","regexp":{"exp" : "[a-z]+_[0-9]+", "matchNo": 1}}          
            }         
        }
    ]
}
```
### Capture Header Value
```json
{
    "steps": [
        {
            "capture_env": {
                "TOKEN": {"from":"header", "header_key":"Authorization"},
            }         
        }
    ]
}
```

### Scenario-Scoped Variables
```json
{
   "env":{
        "TARGET_URL" : "http://localhost:8084/hello",
        "USER_KEY" : "ABC",
        "COMPANY_NAME" : "Ddosify",
        "RANDOM_COUNTRY" : "{{_randomCountry}}",
        "NUMBERS" : [22,33,10,52] 
    },
}
```



### :hammer: Overall Config and Injection
On array-like captured variables or environment vars, the **rand( )** function can be utilized.
```json
// ddosify_config_correlation.json
{
    "iteration_count": 100,
    "load_type": "linear",
    "duration": 10,
    "steps": [
        {
            "id": 1,
            "url": "{{TARGET_URL}}",
            "method": "POST",
            "headers": {
                "User-Key": "{{USER_KEY}}",
                "Rand-Selected-Num" : "{{rand(NUMBERS)}}"
            },
            "payload" : "{{COMPANY_NAME}}",
            "capture_env": {
                "NUM" :{"from":"body","json_path":"num"},
                "NAME" :{"from":"body","json_path":"name"},
                "SQUAD" :{"from":"body","json_path":"squad"},
                "PLAYERS" :{"from":"body","json_path":"squad.players"},
                "MESSI" : {"from":"body","json_path":"squad.players.0"},
                "TOKEN" :{"from":"header", "header_key":"Authorization"},
                "CONTENT_TYPE" :{"from":"header", "header_key":"Content-Type" ,"regexp":{"exp":"application\/(\\w)+","matchNo":0}}             
            }         
        },
        {
            "id": 2,
            "url": "{{TARGET_URL}}",
            "method": "POST",
            "headers": {
                "User-Key": "{{USER_KEY}}",
                "Authorization": "{{TOKEN}}",
                "Content-Type" : "{{CONTENT_TYPE}}"
            },
            "payload_file" : "payload.json",
            "capture_env": {
                "TITLE" :{"from":"body","xpath":"//item/title"},
                "REGEX_MATCH_ENV" :{"from":"body","regexp":{"exp" : "[a-z]+_[0-9]+", "matchNo": 1}}
            }
        }
    ],
    "env":{
        "TARGET_URL" : "http://localhost:8084/hello",
        "USER_KEY" : "ABC",
        "COMPANY_NAME" : "Ddosify",
        "RANDOM_COUNTRY" : "{{_randomCountry}}",
        "NUMBERS" : [22,33,10,52] 
    },
}
```
```json
// payload.json
{
    "boolField" : "{{_randomBoolean}}",
    "numField" : "{{NUM}}",
    "strField" : "{{NAME}}",
    "numArrayField" : ["{{NUM}}",34],
    "strArrayField" : ["{{NAME}}","hello"],
    "mixedArrayField" : ["{{NUM}}",34,"{{NAME}}","{{SQUAD}}"],
    "{{NAME}}" : "messi",
    "obj" :{
        "numField" : "{{NUM}}",
        "objectField" : "{{SQUAD}}",
        "arrayField" : "{{PLAYERS}}"
    }
}
```

## Test Data Set
Ddosify enables you to load test data from **CSV** files. Later, in your scenario, you can inject variables that you tagged.

We are using this [CSV data](https://github.com/ddosify/ddosify/tree/master/config/config_testdata/test.csv) in below config.


```json
// config_data_csv.json
"data":{
      "csv_test": {
          "path" : "config/config_testdata/test.csv",
          "delimiter": ";",
          "vars": {
                  "0":{"tag":"name"},
                  "1":{"tag":"city"},
                  "2":{"tag":"team"},
                  "3":{"tag":"payload", "type":"json"},
                  "4":{"tag":"age", "type":"int"}
                },
          "allow_quota" : true,
          "order": "random",
          "skip_first_line" : true
      }
    }
```

You can refer to tagged variables in your request like below.

```json
// payload.json
{
    "name" : "{{data.csv_test.name}}",
    "team" : "{{data.csv_test.team}}",
    "city" : "{{data.csv_test.city}}",
    "payload" : "{{data.csv_test.payload}}",
    "age" : "{{data.csv_test.age}}"
}
```

## Cookies

Ddosify supports cookies in the following engine modes, `distinct-user` and `repeated-user`. Cookies are not supported in default `ddosify` mode.

In `repeated-user` mode Ddosify uses the same cookie jar for all iterations executed by the same user. It sets cookies returned at first successful iteration and does not change them afterwards. This way same cookies are passed through steps in all iterations executed by the same user.

In `distinct-user` mode Ddosify uses a different cookie jar for each iteration, cookies passed through steps in one iteration only.

You can see an cookie example in [EXAMPLES](https://github.com/ddosify/ddosify/blob/master/engine_docs/EXAMPLES.md#example-1-cookie-support) file.

### Initial / Custom Cookies

You can set initial/custom cookies for your test scenario using `cookie_jar` field in the config file. You can enable/disable custom cookies with `enabled` key. Check the [example config](https://github.com/ddosify/ddosify/tree/master/config/config_testdata/config_init_cookies.json).


| Key       | Description                                                                                                     | Example                                                         |
|-----------|-----------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------|
| `name`      | The name of the cookie. This field is used to identify the cookie.                                              | `platform`                                                        |
| `value`     | The value of the cookie. This field contains the data that the cookie stores.                                   | `web`                                                             |
| `domain`    | Domain or subdomain that can access the cookie.                                                                 | `httpbin.ddosify.com`                                             |
| `path`      | Path within the domain that can access the cookie.                                                              | `/`                                                               |
| `expires`   | When the cookie should expire. The date format should be rfc2616.                                               | `Thu, 16 Mar 2023 09:24:02 GMT`                                   |
| `max_age`   | Number of seconds until the cookie expires.                                                                     | `5`                                                               |
| `http_only` | Whether the cookie should only be accessible through HTTP or HTTPS headers, and not through client-side scripts | `true`                                                            |
| `secure`    | Whether the cookie should only be sent over a secure (HTTPS) connection                                         | `false`                                                            |
| `raw`       | The raw format of the cookie. If it is used, the other keys are discarded.                                      | `myCookie=myValue; Expires=Wed, 21 Oct 2026 07:28:00 GMT; Path=/` |


### Cookie Capture
You can capture values from cookies from its name just like you do for headers and body and use them in your test scenario.

```json
{
    "iteration_count": 100,
    "load_type": "linear",
    "duration": 10,
    "steps": [
        {
          ...
          "capture_env": {
            "TEST" :{"from":"cookies","cookie_name":"test"}
          }
        }
    ]
}
```



### Cookie Assertion
You can refer to cookie values as `cookies.cookie_name` while you write assertions for your steps.

Following fields are available for cookie assertion:
- `name`: Name of the cookie
- `domain`: Domain of the cookie
- `path`: Path of the cookie
- `value`: Value of the cookie
- `expires`: Expiration date of the cookie
- `maxAge`: Max age of the cookie
- `secure`: Secure flag of the cookie
- `httpOnly`: Http only flag of the cookie
- `rawExpires`: Raw expiration date of the cookie

**Examples:**
- `cookies.test.expires < time(\"Thu, 01 Jan 1990 00:00:00 GMT\")` is a valid assertion expression. It checks if the cookie named `test` has an expiration date before `Thu, 01 Jan 1990 00:00:00 GMT`.
- `cookies.test.path == \"/login\"` is another valid assertion expression. It checks if the cookie named `test` has a path value equal to `/login`.

## Tutorials / Blog Posts

* [Testing the Performance of User Authentication Flow](https://ddosify.com/blog/testing-the-performance-of-user-authentication-flow#introduction)
* [Load Testing a Fintech API with CSV Test Data Import](https://ddosify.com/blog/load-testing-a-fintech-exchange-api-with-csv-test-data-import)

## Common Issues

### macOS Security Issue

```
"ddosify" can’t be opened because Apple cannot check it for malicious software.
```

- Open `/usr/local/bin`
- Right click `ddosify` and select Open
- Select Open
- Close the opened terminal

## Communication

You can join our [Discord Server](https://discord.gg/9KdnrSUZQg) for issues, feature requests, feedbacks or anything else.

## More

This repository includes the single-node version of the Ddosify Loader. For distributed and Geo-targeted Load Testing you can use [Ddosify Cloud](https://ddosify.com)

## Disclaimer

Ddosify is created for testing the performance of web applications. Users must be the owner of the target system. Using it for harmful purposes is extremely forbidden. Ddosify team & company is not responsible for its’ usages and consequences.

## License

Licensed under the AGPLv3: https://www.gnu.org/licenses/agpl-3.0.html