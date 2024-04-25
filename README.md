## File Health Analysis with ClamAV and Go

Usage
======


Set Environment Variables: 
- CLAMD_HOST = localhost
- CLAMD_PORT = 3310
- LISTEN_PORT = 8080



`docker run --name clamav -v ${PWD}/clamdata:/var/lib/clamav -d -p 3310:3310 clamav/clamav-debian:latest`

`docker run -e CLAMD_HOST=clamav --link=clamav -p 8080:8080 rscheele3214/scanner:latest`



API:
----

/ and /healthz  Both run a clamd.Ping and return OK if ClamD is contactable, error if not. 
Example:
```
curl -i http://localhost:8080/
HTTP/1.1 200 OK
Content-Type: application/json; charset=UTF-8
Vary: Origin
Date: Fri, 02 Sep 2022 09:54:30 GMT
Content-Length: 5

"OK"
```

error: 

```
curl -i http://localhost:8080/
HTTP/1.1 500 Internal Server Error
Content-Type: application/json; charset=UTF-8
Vary: Origin
Date: Fri, 02 Sep 2022 10:01:15 GMT
Content-Length: 35

{"message":"Could not ping clamd"}
```


Scanning Files
--------------



`/scan` returns a JSON object with the information of what was found in Raw and Description fields.

```
curl -i -F "name=eicar" -F "file=@./eicar.com" http://localhost:8080/scan
HTTP/1.1 451 Unavailable For Legal Reasons
Content-Type: application/json; charset=UTF-8
Vary: Origin
Date: Fri, 02 Sep 2022 10:09:45 GMT
Content-Length: 134

{"Raw":"stream: Win.Test.EICAR_HDB-1 FOUND","Description":"Win.Test.EICAR_HDB-1","Path":"stream","Hash":"","Size":0,"Status":"FOUND"}
```

Clean Files:
```
curl -i -F "name=kobelogs" -F "file=@/Users/Abdulrahman/scanner/kobelogs.tar" http://localhost:8080/scan
HTTP/1.1 100 Continue

HTTP/1.1 200 OK
Content-Type: application/json
Vary: Origin
Date: Thu, 25 Apr 2024 04:20:05 GMT
Content-Length: 87

{"Raw":"stream: OK","Description":"","Path":"stream","Hash":"","Size":0,"Status":"OK"}
```

