For Ubuntu/Debian (should be for all LTS)

```bash
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 379CE192D401AB61
echo "deb https://dl.bintray.com/reddec/debian all main" | sudo tee /etc/apt/sources.list.d/trusted-cgi.list
sudo apt update
sudo apt install trusted-cgi
```

Optionally you may install a server and a client separately.


**Server only**

Daemon only.

```bash
sudo apt install trusted-cgi-server
```

**Client only**

CLI tools only.

```bash
sudo apt install trusted-cgi-client
```