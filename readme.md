Short URL expander implement by API gateway, Lambda and DynamoDB

Project run command

1. 如果沒有安裝過 Serverless framework
```
npm install -g serverless
```
2. 編譯
```
make build
```
3. 部署
```
sls deploy
```
部署特定函式
```
sls deploy function -f functionName
```