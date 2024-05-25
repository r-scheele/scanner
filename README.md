# File Health Analysis with ClamAV and Go

Testing:


Start ClamAV and Scanner:

```bash
docker network create clamav_net
```

```bash
docker run --name clamav -d -p 3310:3310 --network clamav_net clamav/clamav-debian:latest
```

```bash
docker run -e GOOGLE_APPLICATION_CREDENTIALS="/config/peak-essence-171622-ed77209baf22.json" -e CLAMD_HOST=clamav -e BUCKET_NAME="subnet-filescan-test" -e SUBNET_ENDPOINT="http://localhost:8081" -e CLAM_ADDRESS=tcp://clamav:3310 -p 8080:8080 --network clamav_net -v ${PWD}/config:/config rscheele3214/scanner:latest
```


```json
curl -X POST http://localhost:8080/scan/path -H "Content-Type: application/json" -d '{
    "filePath": "clamav-eicar"
}'
```


```json
curl -X POST http://localhost:8080/scan/path -H "Content-Type: application/json" -d '{
    "filePath": "kobelogs.tar"
}'
[{"Raw":"stream: OK","Description":"","Path":"stream","Hash":"","Size":0,"Status":"OK"}]
```

```json
curl -X POST http://localhost:8080/scan/paths -H "Content-Type: application/json" -d '{
    "filePaths": ["kobelogs.tar", "clamav-eicar"]
}'


[{"Raw":"stream: Eicar-Signature FOUND","Description":"Eicar-Signature","Path":"stream","Hash":"","Size":0,"Status":"FOUND"},{"Raw":"stream: OK","Description":"","Path":"stream","Hash":"","Size":0,"Status":"OK"}]
```




To scan a single, non-infected file, use the following `curl` command:

```bash
curl -i -F "name=clamav-eicar" -F "file=@./kobelogs.tar" http://localhost:8080/scan/file
```

**Expected Response:**
```
HTTP/1.1 200 OK
Content-Type: application/json
Date: Thu, 16 May 2024 11:31:01 GMT
Content-Length: 89

[{"Raw":"stream: OK","Description":"","Path":"stream","Hash":"","Size":0,"Status":"OK"}]
```

#### Multiple File Scan

To scan multiple files at once, use this `curl` command:

```bash
curl -i -F "name=clamav-eicar" -F "file=@./kobelogs.tar" -F "file=@./clamav-eicar" http://localhost:8080/scan/files
```

**Expected Response:**
```
HTTP/1.1 200 OK
Content-Type: application/json
Date: Thu, 16 May 2024 11:31:15 GMT
Content-Length: 213

[{"Raw":"stream: OK","Description":"","Path":"stream","Hash":"","Size":0,"Status":"OK"}, {"Raw":"stream: Eicar-Signature FOUND","Description":"Eicar-Signature","Path":"stream","Hash":"","Size":0,"Status":"FOUND"}]
```

### Scanning Potentially Infected Files

To scan a file containing an Eicar test signature or another test file that simulates a virus, use the following `curl` command:

```bash
curl -i -F "name=clamav-eicar" -F "file=@./clamav-eicar" http://localhost:8080/scan/file
```

**Expected Response:**
```
HTTP/1.1 200 OK
Content-Type: application/json
Date: Thu, 16 May 2024 11:31:59 GMT
Content-Length: 126

[{"Raw":"stream: Eicar-Signature FOUND","Description":"Eicar-Signature","Path":"stream","Hash":"","Size":0,"Status":"FOUND"}]
```

### General Information

- **URL**: Replace `localhost:8080` with your server's URL and port.
- **File Path**: Replace `./kobelogs.tar`, `./clamav-eicar` with the path to the file you wish to scan.
