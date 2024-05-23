# File Health Analysis with ClamAV and Go

Testing:

Set Environment Variables: 

`export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service_account.json`
`export CLAMD_HOST=localhost` \
`export CLAMD_PORT=3310`   \
`export LISTEN_PORT=8080`


Start ClamAV and Scanner:

```bash
docker network create clamav_net
```

```bash
docker run --name clamav -v ${PWD}/clamdata:/var/lib/clamav -d -p 3310:3310 --network clamav_net clamav/clamav-debian:latest
```

```bash
docker run -e CLAMD_HOST=clamav -p 8080:8080 --network clamav_net rscheele3214/scanner:latest
```


```json
curl -X POST http://localhost:8080/scan/path -H "Content-Type: application/json" -d '{
    "filePath": "clamav-eicar",
    "subnetEndpoint": "http://example.com/post-results"
}'
```
```json
curl -X POST http://localhost:8080/scan/path -H "Content-Type: application/json" -d '{
    "filePath": "kobelogs.tar",
    "subnetEndpoint": "http://example.com/post-results"
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
    "filePaths": ["kobelogs.tar", "clamav-eicar"],
    "subnetEndpoint": "http://example.com/receive_scan_results"
}'


[{"Raw":"stream: Eicar-Signature FOUND","Description":"Eicar-Signature","Path":"stream","Hash":"","Size":0,"Status":"FOUND"},{"Raw":"stream: OK","Description":"","Path":"stream","Hash":"","Size":0,"Status":"OK"}]
```

Scalability:

Adjust the StreamMaxLength in the clamd.conf file to allow for larger files to be scanned.

1. **Copy the Configuration File**:
   ```bash
   docker cp clamav:/etc/clamav/clamd.conf ./config
   ```

2. **Edit Locally**: Edit the copied `clamd.conf` file to change the `StreamMaxLength` value to a higher value.

3. **Copy Back to Container**: 
   ```bash
   docker cp ./config/clamd.conf clamav:/etc/clamav/clamd.conf
   ```

4. **Restart ClamAV**: 
   ```bash
   docker restart clamav
   ```

### Scanning Non-Infected Files

#### Single File Scan

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
