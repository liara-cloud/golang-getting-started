# golang-getting-started
first steps with Golang Projects in Liara

## Availabe Branches

1.  [Disk setup](https://github.com/liara-cloud/golang-getting-started/tree/diskSetup)
2.  [Object Storage In Liara](https://github.com/liara-cloud/golang-getting-started/tree/upload-using-s3)
3.  [Dockerized Go For Deploying](https://github.com/liara-cloud/golang-getting-started/tree/go-dockerized)
4.  [Upload File Using S3 In Liara](https://github.com/liara-cloud/golang-getting-started/tree/upload-using-s3)

## Blog Website For Test
## Installation

```bash
  git clone https://github.com/liara-cloud/golang-getting-started.git
```
```bash
  cd golang-getting-started
```
```bash
  cp .env.example .env
```
- or if you're using windows, just rename .env.example to .env
- configure your environment variables
```bash
  mkdir static/images
```
- or if you're using windows, create image folder in static folder
```bash
  go run main.go
```
- if you need to Live reload, you can use air:
```bash
  go install github.com/cosmtrek/air@latest
```
```bash
  air init
```
```bash
  air
```

## Documentation
Read more on liara [Golang apps documentation](https://docs.liara.ir/app-deploy/golang/getting-started/)
