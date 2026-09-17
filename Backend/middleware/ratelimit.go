package middleware

import (
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Limita tentativas de login/cadastro por IP, para dificultar força bruta
// (adivinhar senha por tentativa e erro) e criação em massa de contas falsas
// agora que o app é público.
//
// Só conta tentativas que FALHARAM (status >= 400). Um login certo, mesmo
// repetido várias vezes seguidas, nunca é bloqueado — só erros consecutivos.

const (
	maxTentativas      = 5
	janelaDeTentativas = 5 * time.Minute
)

type registroTentativas struct {
	mu         sync.Mutex
	tentativas map[string][]time.Time
}

var tentativasPorIP = &registroTentativas{
	tentativas: make(map[string][]time.Time),
}

// LimitarTentativas envolve um handler e bloqueia com 429 se o IP já tiver
// atingido o limite de tentativas falhas na janela de tempo.
func LimitarTentativas(proximo http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := obterIP(r)

		if bloqueado(ip) {
			w.Header().Set("Retry-After", "300")
			http.Error(w, "muitas tentativas, tente novamente em alguns minutos", http.StatusTooManyRequests)
			return
		}

		// Precisamos capturar o status de resposta do handler original para
		// saber se a tentativa falhou ou teve sucesso — http.ResponseWriter
		// não expõe isso por padrão, então usamos um wrapper simples.
		captura := &capturadorDeStatus{ResponseWriter: w, status: http.StatusOK}
		proximo(captura, r)

		if captura.status >= 400 {
			registrarFalha(ip)
		} else {
			// Login/cadastro deu certo: limpa o histórico desse IP, para não
			// acumular tentativas antigas que não representam mais um risco.
			limparIP(ip)
		}
	}
}

func bloqueado(ip string) bool {
	tentativasPorIP.mu.Lock()
	defer tentativasPorIP.mu.Unlock()
	limparRegistrosExpirados()

	tentativas := limpasAntigasTentativas(tentativasPorIP.tentativas[ip])
	tentativasPorIP.tentativas[ip] = tentativas

	return len(tentativas) >= maxTentativas
}

func registrarFalha(ip string) {
	tentativasPorIP.mu.Lock()
	defer tentativasPorIP.mu.Unlock()

	tentativas := limpasAntigasTentativas(tentativasPorIP.tentativas[ip])
	tentativasPorIP.tentativas[ip] = append(tentativas, time.Now())
}

func limparIP(ip string) {
	tentativasPorIP.mu.Lock()
	defer tentativasPorIP.mu.Unlock()
	delete(tentativasPorIP.tentativas, ip)
}

// limpasAntigasTentativas remove tentativas fora da janela de tempo, para o
// mapa não crescer para sempre e para tentativas antigas não contarem contra
// o limite indefinidamente.
func limpasAntigasTentativas(tentativas []time.Time) []time.Time {
	limite := time.Now().Add(-janelaDeTentativas)
	restantes := tentativas[:0]
	for _, t := range tentativas {
		if t.After(limite) {
			restantes = append(restantes, t)
		}
	}
	return restantes
}

// obterIP tenta pegar o IP real do cliente atrás do Cloudflare Tunnel/nginx
// (que colocam o IP original em cabeçalhos), com fallback pro IP da conexão direta.
func obterIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}
	if net.ParseIP(ip) == nil {
		return "ip-invalido"
	}

	// Cabeçalhos de proxy só são confiáveis quando a conexão veio de uma rede
	// explicitamente configurada. Isso evita que uma chamada direta ao backend
	// falsifique CF-Connecting-IP/X-Forwarded-For para contornar o rate limit.
	if proxyConfiavel(ip) {
		if cf := net.ParseIP(strings.TrimSpace(r.Header.Get("CF-Connecting-IP"))); cf != nil {
			return cf.String()
		}
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			partes := strings.Split(xff, ",")
			if cliente := net.ParseIP(strings.TrimSpace(partes[0])); cliente != nil {
				return cliente.String()
			}
		}
	}

	return net.ParseIP(ip).String()
}

// proxyConfiavel verifica o IP do peer contra TRUSTED_PROXY_CIDRS, uma lista
// separada por vírgulas (por exemplo: "127.0.0.1/32,172.18.0.0/16"). Sem a
// configuração, nenhum cabeçalho de proxy é aceito.
func proxyConfiavel(ip string) bool {
	redes := strings.Split(os.Getenv("TRUSTED_PROXY_CIDRS"), ",")
	endereco := net.ParseIP(ip)
	for _, textoRede := range redes {
		_, rede, err := net.ParseCIDR(strings.TrimSpace(textoRede))
		if err == nil && rede.Contains(endereco) {
			return true
		}
	}
	return false
}

func limparRegistrosExpirados() {
	if len(tentativasPorIP.tentativas) <= 10000 {
		return
	}
	for ip, tentativas := range tentativasPorIP.tentativas {
		if len(limpasAntigasTentativas(tentativas)) == 0 {
			delete(tentativasPorIP.tentativas, ip)
		}
	}
}

// capturadorDeStatus é um wrapper mínimo de http.ResponseWriter só para
// guardar qual status code o handler original respondeu.
type capturadorDeStatus struct {
	http.ResponseWriter
	status int
}

func (c *capturadorDeStatus) WriteHeader(status int) {
	c.status = status
	c.ResponseWriter.WriteHeader(status)
}
