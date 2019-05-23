# makeSnapshot

makeSnapshot is a CLI tool to create a snapshot of a virtual machine on the KPN vRA platform.
Only one snapshot per VM is allowed. The default behaviour is to overwrite the existing snapshot.

Several API calls are needed to creates a snapshot of the virtual machine.

Tracing can be turned on to provide information on the progress.

After the snapshot request is send the status of the request is checked every 10 seconds.
The time between request and the final status can take half-a-minute or more.

Required parameters like the baseURL, tenant, domain, and credentials are read from a 'yaml' config file

```yaml
---
baseURL: "https://base-platformURL"
tenant: "tenantName"
domain: "login domain"
userName: "userName"
password: "password"
...
```

When the snapshot is created the app exits with status code 0.
On failure or error the app exits with status code 1.
The exit code is not displayed but can be checked at the commandline with `echo $?`

DISCLAIMER:
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.

## Build

The software is developed on a Mac. The Linux version was created with the following command:

`env GOOS=linux GOARCH=386 go build .`
