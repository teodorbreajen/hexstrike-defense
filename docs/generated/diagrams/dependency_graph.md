```mermaid
flowchart TB
    subgraph Dependencies
        v1["corev1 "k8s.io/api/core/v1"]
        go["dto "github.com/prometheus/client_model/go"]
        v5["github.com/golang-jwt/jwt/v5"]
        uuid["github.com/google/uuid"]
        framework["github.com/hexstrike/hexstrike-defense/tests/e2e/framework"]
        dlq["github.com/hexstrike/mcp-policy-proxy/dlq"]
        prometheus["github.com/prometheus/client_golang/prometheus"]
        promauto["github.com/prometheus/client_golang/prometheus/promauto"]
        promhttp["github.com/prometheus/client_golang/prometheus/promhttp"]
        assert["github.com/stretchr/testify/assert"]
        require["github.com/stretchr/testify/require"]
        kubernetes["k8s.io/client-go/kubernetes"]
        scheme["k8s.io/client-go/kubernetes/scheme"]
        clientcmd["k8s.io/client-go/tools/clientcmd"]
        remotecommand["k8s.io/client-go/tools/remotecommand"]
        v1["metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"]
        rest["restclient "k8s.io/client-go/rest"]
        sort["sort"]
        strconv["strconv"]
        syscall["syscall"]
        testing["testing"]
    end
```