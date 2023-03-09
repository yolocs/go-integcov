# go-integcov

1. Use `ghcr.io/yolocs/go-integcov` as the base image for your Go app in
   integration test. The image was built from
   [`cgr.dev/chainguard/static`](https://edu.chainguard.dev/chainguard/chainguard-images/reference/static/overview/)
   with the `go-integcov` being the only additional binary.

2. When building your Go app for integration test, add flag `-cover`. See
   https://go.dev/testing/coverage/.

2. When starting you Go app in integration test, use `go-integcov` to wrap your
   program. E.g. `go-integcov my-app -flag1 -flag2`.

3. Set env var `INTEGCOV_STORAGE` to `gs://a-gcs-bucket-you-own/a-folder`.

4. After you integration test is completed, copy all the files from
   `gs://a-gcs-bucket-you-own/a-folder` and use `go tool covdata` to understand
   your integration test coverage. See https://go.dev/testing/coverage/.
