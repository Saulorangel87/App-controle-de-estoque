import { useEffect, useState } from "react";
import {
  obterEmailConta,
  solicitarEmailConta,
  verificarEmailConta,
} from "../api/api.js";

export default function ModalEmailConta({ token, aoFechar }) {
  const [email, setEmail] = useState("");
  const [codigo, setCodigo] = useState("");
  const [verificado, setVerificado] = useState(false);
  const [aguardandoCodigo, setAguardandoCodigo] = useState(false);
  const [carregando, setCarregando] = useState(true);
  const [erro, setErro] = useState("");
  const [sucesso, setSucesso] = useState("");

  useEffect(() => {
    obterEmailConta(token)
      .then((conta) => {
        setEmail(conta.email ?? "");
        setVerificado(Boolean(conta.verificado));
      })
      .catch(() => setErro("Não foi possível carregar o e-mail da conta."))
      .finally(() => setCarregando(false));
  }, [token]);

  async function enviarCodigo(evento) {
    evento.preventDefault();
    setErro("");
    setSucesso("");
    setCarregando(true);
    try {
      const resposta = await solicitarEmailConta(email, token);
      if (resposta.mensagem === "este e-mail já está verificado") {
        setVerificado(true);
        setAguardandoCodigo(false);
        setSucesso(resposta.mensagem);
        return;
      }
      setAguardandoCodigo(true);
      setVerificado(false);
      setSucesso(resposta.mensagem || "Código enviado para o e-mail informado.");
    } catch (e) {
      setErro(
        e.status === 503
          ? "O envio de e-mail está temporariamente indisponível."
          : e.status === 409
          ? "Esse e-mail já está em uso."
          : "Não foi possível enviar o código."
      );
    } finally {
      setCarregando(false);
    }
  }

  async function confirmarCodigo(evento) {
    evento.preventDefault();
    setErro("");
    setSucesso("");
    setCarregando(true);
    try {
      await verificarEmailConta(codigo, token);
      setVerificado(true);
      setAguardandoCodigo(false);
      setCodigo("");
      setSucesso("E-mail confirmado com sucesso.");
    } catch {
      setErro("Código inválido ou expirado.");
    } finally {
      setCarregando(false);
    }
  }

  return (
    <div
      className="modal-fundo"
      role="dialog"
      aria-modal="true"
      aria-labelledby="titulo-modal-email"
      onClick={(e) => e.target === e.currentTarget && aoFechar()}
    >
      <div className="modal-caixa">
        <h2 id="titulo-modal-email">E-mail da conta</h2>
        <p className="subtitulo">
          Confirme um e-mail para recuperar sua senha com segurança.
        </p>

        {erro && <p className="mensagem-erro" role="alert">{erro}</p>}
        {sucesso && (
          <p className="mensagem-erro" role="status" style={{ color: "var(--cor-primaria)" }}>
            {sucesso}
          </p>
        )}

        {carregando && !email ? (
          <p className="estado-vazio">Carregando...</p>
        ) : (
          <>
            <form onSubmit={enviarCodigo} noValidate>
              <div className="campo-formulario">
                <label htmlFor="campo-email-conta">E-mail</label>
                <input
                  id="campo-email-conta"
                  type="email"
                  autoComplete="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                />
              </div>
              <button type="submit" className="botao botao-primario" disabled={carregando}>
                {verificado ? "Trocar e-mail" : "Enviar código"}
              </button>
              {verificado && <span className="subtitulo"> E-mail verificado</span>}
            </form>

            {aguardandoCodigo && (
              <form onSubmit={confirmarCodigo} noValidate>
                <div className="campo-formulario" style={{ marginTop: "var(--espaco-4)" }}>
                  <label htmlFor="campo-codigo-conta">Código recebido</label>
                  <input
                    id="campo-codigo-conta"
                    type="text"
                    inputMode="numeric"
                    autoComplete="one-time-code"
                    maxLength={8}
                    required
                    value={codigo}
                    onChange={(e) => setCodigo(e.target.value.replace(/\D/g, ""))}
                  />
                </div>
                <button type="submit" className="botao botao-primario" disabled={carregando}>
                  {carregando ? "Confirmando..." : "Confirmar e-mail"}
                </button>
              </form>
            )}
          </>
        )}

        <div className="acoes-modal">
          <button type="button" className="botao botao-secundario" onClick={aoFechar}>
            Fechar
          </button>
        </div>
      </div>
    </div>
  );
}
