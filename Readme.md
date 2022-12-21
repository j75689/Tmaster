# Tmaster

<p>
  <img src="https://img.shields.io/github/actions/workflow/status/j75689/tmaster/checker.yml?branch=main">
  <a href="https://github.com/j75689/Tmaster/blob/master/LICENSE">
    <img src="https://img.shields.io/github/license/globocom/go-buffer?color=blue&style=flat-square">
  </a>
  <img src="https://img.shields.io/github/go-mod/go-version/j75689/Tmaster?style=flat-square">
  <a href="https://pkg.go.dev/github.com/j75689/Tmaster">
    <img src="https://img.shields.io/badge/Go-reference-blue?style=flat-square">
  </a>
</p>

This project was born from the inspiration of `AWS Step Function`.

It can be used to orchestrate services and build serverless applications. Workflows manage failures, retries, parallelization, service integrations, and observability so developers can focus on higher-value business logic.

## Infrastructure

![infrastructure](infra.png)


## Examples

```graphql
mutation begin{
  CreateJob(input: {
    comment: "demo"
    parameters: "{\"xxx\":{\"body\":\"abc\"},\"max_retry\": 1}"
    start_at: "test_grpc"
    tasks: [
      {
        name:"open_google_search"
        type: TASK
        next: "test_grpc"
        result_path: "${result.text}"
        endpoint: {
          protocol: HTTP
          method: GET
          url: "https://www.google.com/?hl=zh-tw"
          body: "${body}"
        },
        timeout: 10,
        retry: {
          error_on: [TIMEOUT]
          max_attempts: "${max_retry}"
          interval: 1
        }
	catch: {
          error_on: [ALL],
          next: "open_baidu_search"
        }
      },
      {
        name:"open_baidu_search",
        type: TASK,
        next: "test_grpc"
        endpoint: {
          protocol: HTTP
          method: GET
          url: "https://www.baidu.com/"
        },
        timeout: 60,
        retry: {
          error_on: [TIMEOUT]
          max_attempts: "${max_retry}"
          interval: 1
        }
      },
      {
        name: "test_grpc"
        type: TASK
        endpoint: {
          protocol: GRPC
          url: "example-grpc:10050"
          body:"#{Execution.CauseError}"
          symbol: "example.HealthService/Health"
        }
        result_path: "${result.getitem}"
        output_path: "${result.getitem}"
        end: true
      }
    ]
  })
}
```
