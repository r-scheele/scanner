# File Health Analysis with ClamAV and Go

Testing:

Set Environment Variables: 

- CLAMD_HOST = localhost
- CLAMD_PORT = 3310
- LISTEN_PORT = 8080


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




API:

The scanner provides the following endpoints for scanning files:

1. **Ping Handler**:

   - **Method**: GET
   - **URL**: `/ping`

2. **Scan Stream Handler**:

   - **Method**: POST
   - **URL**: `/scan/stream`
   - **Body**: Stream data to be scanned

   `curl -i -F "name=kobelogs" -F "file=@/Users/Abdulrahman/scanner/kobelogs.tar" http://localhost:8080/scan/stream`

    ```
    HTTP/1.1 100 Continue

    HTTP/1.1 200 OK
    Content-Type: application/json
    Date: Fri, 03 May 2024 15:11:47 GMT
    Content-Length: 89

    [{"Raw":"stream: OK","Description":"","Path":"stream","Hash":"","Size":0,"Status":"OK"}]
    ```

3. **Scan File Handler**:

   - **Method**: POST
   - **URL**: `/scan/file`
   - **Body**: `{"path": "/path/to/file"}`

4. **Scan Files Handler**:

   - **Method**: POST
   - **URL**: `/scan/files`
   - **Body**: `["/path/to/file1", "/path/to/file2", "/path/to/file3"]`


Scalability:

Adjust the StreamMaxLength in the clamd.conf file to allow for larger files to be scanned.

1. **Copy the Configuration File**:
   ```bash
   docker cp clamav:/etc/clamav/clamd.conf .
   ```

2. **Edit Locally**: Edit the copied `clamd.conf` file on your local machine using your preferred text editor.

3. **Copy Back to Container**: 
   ```bash
   docker cp clamd.conf clamav:/etc/clamav/clamd.conf
   ```

4. **Restart ClamAV**: 
   ```bash
   docker restart clamav
   ```
