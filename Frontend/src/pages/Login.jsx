import { useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  login,
  cadastrar,
  verificarEmailCadastro,
  reenviarVerificacaoEmail,
  solicitarRecuperacaoSenha,
  redefinirSenha,
} from "../api/api.js";
import { useAuth } from "../context/AuthContext.jsx";

const MODOS = {
  ENTRAR: "entrar",
  CADASTRO: "cadastro",
  CADASTRO_VERIFICAR: "cadastro-verificar",
  RECUPERAR_EMAIL: "recuperar-email",
  RECUPERAR_REDEFINIR: "recuperar-redefinir",
};

export default function Login() {
  const [modo, setModo] = useState(MODOS.ENTRAR);
  const [nome, setNome] = useState("");
  const [email, setEmail] = useState("");
  const [senha, setSenha] = useState("");
  const [codigo, setCodigo] = useState("");
  const [novaSenha, setNovaSenha] = useState("");
  const [erro, setErro] = useState("");
  const [sucesso, setSucesso] = useState("");
  const [carregando, setCarregando] = useState(false);

  const { entrar } = useAuth();
  const navegar = useNavigate();

  function irParaModo(novoModo) {
    setModo(novoModo);
    setErro("");
    setSucesso("");
    setCodigo("");
  }

  function mensagemRateLimit(status) {
    if (status === 429) {
      return "Muitas tentativas seguidas. Aguarde alguns minutos antes de tentar de novo.";
    }
    return "";
  }

  async function aoEnviarEntrarOuCadastrar(evento) {
    evento.preventDefault();
    setErro("");
    setCarregando(true);

    try {
      if (modo === MODOS.CADASTRO) {
        await cadastrar(nome, email, senha);
        setModo(MODOS.CADASTRO_VERIFICAR);
        setSucesso("Enviamos um código de confirmação para seu e-mail.");
        return;
      }

      const resultado = await login(nome, senha);
      entrar(resultado.nome);
      navegar("/");
    } catch (e) {
      const limite = mensagemRateLimit(e.status);
      setErro(
        limite ||
          (modo === MODOS.ENTRAR && e.status === 403
            ? "Confirme seu e-mail antes de entrar."
            : modo === MODOS.CADASTRO
            ? e.status === 503
              ? "O envio de e-mail está temporariamente indisponível. Tente novamente mais tarde."
              : e.status === 409
              ? "Esse nome de usuário ou e-mail já está cadastrado."
              : "Não foi possível criar a conta. Confira os dados informados."
            : "Usuário ou senha inválidos.")
      );
    } finally {
      setCarregando(false);
    }
  }

  async function aoVerificarCadastro(evento) {
    evento.preventDefault();
    setErro("");
    setCarregando(true);
    try {
      await verificarEmailCadastro(nome, codigo);
      const resultado = await login(nome, senha);
      entrar(resultado.nome);
      navegar("/");
    } catch (e) {
      setErro(
        mensagemRateLimit(e.status) || "Código inválido ou expirado. Solicite um novo código."
      );
    } finally {
      setCarregando(false);
    }
  }

  async function aoReenviarVerificacao() {
    setErro("");
    setSucesso("");
    setCarregando(true);
    try {
      await reenviarVerificacaoEmail(nome);
      setSucesso("Se a conta estiver pendente, um novo código será enviado em breve.");
    } catch (e) {
      setErro(
        mensagemRateLimit(e.status) ||
          "Não foi possível reenviar o código. Tente novamente mais tarde."
      );
    } finally {
      setCarregando(false);
    }
  }

  async function aoSolicitarRecuperacao(evento) {
    evento.preventDefault();
    setErro("");
    setCarregando(true);
    try {
      await solicitarRecuperacaoSenha(email);
      setModo(MODOS.RECUPERAR_REDEFINIR);
      setSucesso("Se o e-mail estiver cadastrado, você receberá um código em breve.");
    } catch (e) {
      setErro(
        mensagemRateLimit(e.status) ||
          (e.status === 503
            ? "O envio de e-mail está temporariamente indisponível."
            : "Informe um e-mail válido para continuar.")
      );
    } finally {
      setCarregando(false);
    }
  }

  async function aoRedefinirSenha(evento) {
    evento.preventDefault();
    setErro("");
    setCarregando(true);
    try {
      await redefinirSenha(email, codigo, novaSenha);
      setSenha("");
      setCodigo("");
      setNovaSenha("");
      irParaModo(MODOS.ENTRAR);
      setSucesso("Senha redefinida! Já pode entrar com a nova senha.");
    } catch (e) {
      setErro(
        mensagemRateLimit(e.status) || "Código inválido ou expirado. Solicite um novo código."
      );
    } finally {
      setCarregando(false);
    }
  }

  const tituloPorModo = {
    [MODOS.ENTRAR]: "Entrar",
    [MODOS.CADASTRO]: "Criar conta",
    [MODOS.CADASTRO_VERIFICAR]: "Confirmar e-mail",
    [MODOS.RECUPERAR_EMAIL]: "Recuperar senha",
    [MODOS.RECUPERAR_REDEFINIR]: "Recuperar senha",
  };

  return (
    <main className="pagina-login">
      <div className="caixa-login">
        <h1>{tituloPorModo[modo]}</h1>
        <p className="subtitulo">Controle de estoque de mantimentos de casa</p>

        {erro && (
          <p className="mensagem-erro" role="alert">
            {erro}
          </p>
        )}
        {sucesso && (
          <p
            className="mensagem-erro"
            role="status"
            style={{ color: "var(--cor-primaria)", backgroundColor: "transparent", padding: 0 }}
          >
            {sucesso}
          </p>
        )}

        {(modo === MODOS.ENTRAR || modo === MODOS.CADASTRO) && (
          <form onSubmit={aoEnviarEntrarOuCadastrar} noValidate>
            <div className="campo-formulario">
              <label htmlFor="campo-nome">Nome de usuário</label>
              <input
                id="campo-nome"
                name="nome"
                type="text"
                autoComplete="username"
                required
                value={nome}
                onChange={(e) => setNome(e.target.value)}
              />
            </div>

            {modo === MODOS.CADASTRO && (
              <div className="campo-formulario">
                <label htmlFor="campo-email">E-mail</label>
                <input
                  id="campo-email"
                  name="email"
                  type="email"
                  autoComplete="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                />
                <p className="subtitulo" style={{ marginTop: "4px" }}>
                  Usaremos este e-mail para confirmar a conta e recuperar a senha.
                </p>
              </div>
            )}

            <div className="campo-formulario">
              <label htmlFor="campo-senha">Senha</label>
              <input
                id="campo-senha"
                name="senha"
                type="password"
                autoComplete={modo === MODOS.CADASTRO ? "new-password" : "current-password"}
                required
                minLength={modo === MODOS.CADASTRO ? 8 : undefined}
                maxLength={72}
                value={senha}
                onChange={(e) => setSenha(e.target.value)}
              />
            </div>

            <button
              type="submit"
              className="botao botao-primario"
              style={{ width: "100%", justifyContent: "center" }}
              disabled={carregando}
            >
              {carregando ? "Aguarde..." : modo === MODOS.CADASTRO ? "Criar conta" : "Entrar"}
            </button>
          </form>
        )}

        {modo === MODOS.CADASTRO_VERIFICAR && (
          <form onSubmit={aoVerificarCadastro} noValidate>
            <p className="subtitulo">
              Digite o código de 8 números enviado para <strong>{email}</strong>.
            </p>
            <div className="campo-formulario">
              <label htmlFor="campo-codigo-cadastro">Código de confirmação</label>
              <input
                id="campo-codigo-cadastro"
                type="text"
                inputMode="numeric"
                autoComplete="one-time-code"
                pattern="[0-9]{8}"
                maxLength={8}
                required
                value={codigo}
                onChange={(e) => setCodigo(e.target.value.replace(/\D/g, ""))}
              />
            </div>
            <button
              type="submit"
              className="botao botao-primario"
              style={{ width: "100%", justifyContent: "center" }}
              disabled={carregando}
            >
              {carregando ? "Confirmando..." : "Confirmar e-mail"}
            </button>
            <button
              type="button"
              className="alternar-modo"
              onClick={aoReenviarVerificacao}
              disabled={carregando}
            >
              Reenviar código
            </button>
          </form>
        )}

        {modo === MODOS.RECUPERAR_EMAIL && (
          <form onSubmit={aoSolicitarRecuperacao} noValidate>
            <div className="campo-formulario">
              <label htmlFor="campo-email-recuperar">E-mail da conta</label>
              <input
                id="campo-email-recuperar"
                type="email"
                autoComplete="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </div>
            <button
              type="submit"
              className="botao botao-primario"
              style={{ width: "100%", justifyContent: "center" }}
              disabled={carregando}
            >
              {carregando ? "Enviando..." : "Enviar código"}
            </button>
          </form>
        )}

        {modo === MODOS.RECUPERAR_REDEFINIR && (
          <form onSubmit={aoRedefinirSenha} noValidate>
            <p className="subtitulo">
              Informe o código recebido em <strong>{email}</strong>.
            </p>
            <div className="campo-formulario">
              <label htmlFor="campo-codigo-recuperar">Código de recuperação</label>
              <input
                id="campo-codigo-recuperar"
                type="text"
                inputMode="numeric"
                autoComplete="one-time-code"
                pattern="[0-9]{8}"
                maxLength={8}
                required
                value={codigo}
                onChange={(e) => setCodigo(e.target.value.replace(/\D/g, ""))}
              />
            </div>
            <div className="campo-formulario">
              <label htmlFor="campo-nova-senha">Nova senha</label>
              <input
                id="campo-nova-senha"
                type="password"
                autoComplete="new-password"
                required
                minLength={8}
                maxLength={72}
                value={novaSenha}
                onChange={(e) => setNovaSenha(e.target.value)}
              />
            </div>
            <button
              type="submit"
              className="botao botao-primario"
              style={{ width: "100%", justifyContent: "center" }}
              disabled={carregando}
            >
              {carregando ? "Redefinindo..." : "Redefinir senha"}
            </button>
          </form>
        )}

        {modo === MODOS.ENTRAR && (
          <>
            <button type="button" className="alternar-modo" onClick={() => irParaModo(MODOS.CADASTRO)}>
              Ainda não tenho conta — criar
            </button>
            <button
              type="button"
              className="link-recuperar-senha"
              onClick={() => irParaModo(MODOS.RECUPERAR_EMAIL)}
            >
              Esqueci minha senha
            </button>
          </>
        )}

        {(modo === MODOS.CADASTRO || modo === MODOS.CADASTRO_VERIFICAR) && (
          <button type="button" className="alternar-modo" onClick={() => irParaModo(MODOS.ENTRAR)}>
            Já tenho conta — entrar
          </button>
        )}

        {(modo === MODOS.RECUPERAR_EMAIL || modo === MODOS.RECUPERAR_REDEFINIR) && (
          <button type="button" className="alternar-modo" onClick={() => irParaModo(MODOS.ENTRAR)}>
            Voltar para o login
          </button>
        )}
      </div>
    </main>
  );
}
