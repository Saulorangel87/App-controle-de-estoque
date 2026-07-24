import { useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  login,
  cadastrar,
  obterPerguntaSeguranca,
  redefinirSenha,
} from "../api/api.js";
import { useAuth } from "../context/AuthContext.jsx";
import { PERGUNTAS_SEGURANCA } from "../utils/perguntasSeguranca.js";

// Os quatro estados possíveis da tela. Mantê-los como modos dentro do mesmo
// componente (em vez de rotas separadas) evita duplicar o layout do cartão de login.
const MODOS = {
  ENTRAR: "entrar",
  CADASTRO: "cadastro",
  RECUPERAR_PERGUNTA: "recuperar-pergunta", // passo 1: informar o usuário
  RECUPERAR_REDEFINIR: "recuperar-redefinir", // passo 2: responder e criar nova senha
};

export default function Login() {
  const [modo, setModo] = useState(MODOS.ENTRAR);

  const [nome, setNome] = useState("");
  const [senha, setSenha] = useState("");
  const [perguntaSeguranca, setPerguntaSeguranca] = useState(
    PERGUNTAS_SEGURANCA[0]
  );
  const [respostaSeguranca, setRespostaSeguranca] = useState("");

  // Usados só no fluxo de recuperação
  const [perguntaExibida, setPerguntaExibida] = useState("");
  const [respostaRecuperacao, setRespostaRecuperacao] = useState("");
  const [novaSenha, setNovaSenha] = useState("");

  const [erro, setErro] = useState("");
  const [sucesso, setSucesso] = useState("");
  const [carregando, setCarregando] = useState(false);

  const { entrar } = useAuth();
  const navegar = useNavigate();

  // Centraliza a troca de modo para sempre limpar mensagens antigas —
  // evita um erro do formulário anterior "vazar" para a tela seguinte.
  function irParaModo(novoModo) {
    setModo(novoModo);
    setErro("");
    setSucesso("");
  }

  async function aoEnviarEntrarOuCadastrar(evento) {
    evento.preventDefault();
    setErro("");
    setCarregando(true);

    try {
      if (modo === MODOS.CADASTRO) {
        if (!respostaSeguranca.trim()) {
          setErro("Escolha uma pergunta de segurança e informe a resposta.");
          setCarregando(false);
          return;
        }
        // Depois de criar a conta, já faz login automaticamente para não
        // obrigar a pessoa a digitar tudo de novo.
        await cadastrar(nome, senha, perguntaSeguranca, respostaSeguranca);
        const resultado = await login(nome, senha);
        entrar(resultado.token);
      } else {
        const resultado = await login(nome, senha);
        entrar(resultado.token);
      }
      navegar("/");
    } catch (e) {
      setErro(
        modo === MODOS.CADASTRO
          ? "Não foi possível criar a conta. O nome de usuário já pode estar em uso."
          : "Usuário ou senha inválidos."
      );
    } finally {
      setCarregando(false);
    }
  }

  // Passo 1 da recuperação: busca a pergunta de segurança cadastrada para o usuário.
  async function aoBuscarPergunta(evento) {
    evento.preventDefault();
    setErro("");
    setCarregando(true);

    try {
      const resultado = await obterPerguntaSeguranca(nome);
      setPerguntaExibida(resultado.pergunta_seguranca);
      irParaModo(MODOS.RECUPERAR_REDEFINIR);
    } catch (e) {
      setErro("Não foi possível encontrar esse usuário.");
    } finally {
      setCarregando(false);
    }
  }

  // Passo 2 da recuperação: confere a resposta e define a nova senha.
  async function aoRedefinirSenha(evento) {
    evento.preventDefault();
    setErro("");
    setCarregando(true);

    try {
      await redefinirSenha(nome, respostaRecuperacao, novaSenha);
      setSenha("");
      setRespostaRecuperacao("");
      setNovaSenha("");
      irParaModo(MODOS.ENTRAR);
      setSucesso("Senha redefinida! Já pode entrar com a nova senha.");
    } catch (e) {
      setErro("Resposta de segurança incorreta.");
    } finally {
      setCarregando(false);
    }
  }

  const tituloPorModo = {
    [MODOS.ENTRAR]: "Entrar",
    [MODOS.CADASTRO]: "Criar conta",
    [MODOS.RECUPERAR_PERGUNTA]: "Recuperar senha",
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
          <p className="mensagem-erro" role="status" style={{ color: "var(--cor-primaria)", backgroundColor: "transparent", padding: 0 }}>
            {sucesso}
          </p>
        )}

        {/* ---------- Entrar / Criar conta ---------- */}
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

            <div className="campo-formulario">
              <label htmlFor="campo-senha">Senha</label>
              <input
                id="campo-senha"
                name="senha"
                type="password"
                autoComplete={
                  modo === MODOS.CADASTRO ? "new-password" : "current-password"
                }
                required
                minLength={6}
                value={senha}
                onChange={(e) => setSenha(e.target.value)}
              />
            </div>

            {modo === MODOS.CADASTRO && (
              <>
                <div className="campo-formulario">
                  <label htmlFor="campo-pergunta">Pergunta de segurança</label>
                  <select
                    id="campo-pergunta"
                    value={perguntaSeguranca}
                    onChange={(e) => setPerguntaSeguranca(e.target.value)}
                  >
                    {PERGUNTAS_SEGURANCA.map((pergunta) => (
                      <option key={pergunta} value={pergunta}>
                        {pergunta}
                      </option>
                    ))}
                  </select>
                </div>

                <div className="campo-formulario">
                  <label htmlFor="campo-resposta">Resposta</label>
                  <input
                    id="campo-resposta"
                    type="text"
                    required
                    value={respostaSeguranca}
                    onChange={(e) => setRespostaSeguranca(e.target.value)}
                  />
                  <p className="subtitulo" style={{ marginTop: "4px" }}>
                    Você vai precisar dessa resposta se esquecer a senha.
                  </p>
                </div>
              </>
            )}

            <button
              type="submit"
              className="botao botao-primario"
              style={{ width: "100%", justifyContent: "center" }}
              disabled={carregando}
            >
              {carregando
                ? "Aguarde..."
                : modo === MODOS.CADASTRO
                ? "Criar conta"
                : "Entrar"}
            </button>
          </form>
        )}

        {/* ---------- Recuperar senha: passo 1 (informar o usuário) ---------- */}
        {modo === MODOS.RECUPERAR_PERGUNTA && (
          <form onSubmit={aoBuscarPergunta} noValidate>
            <div className="campo-formulario">
              <label htmlFor="campo-nome-recuperar">Nome de usuário</label>
              <input
                id="campo-nome-recuperar"
                type="text"
                autoComplete="username"
                required
                value={nome}
                onChange={(e) => setNome(e.target.value)}
              />
            </div>

            <button
              type="submit"
              className="botao botao-primario"
              style={{ width: "100%", justifyContent: "center" }}
              disabled={carregando}
            >
              {carregando ? "Buscando..." : "Continuar"}
            </button>
          </form>
        )}

        {/* ---------- Recuperar senha: passo 2 (responder e definir nova senha) ---------- */}
        {modo === MODOS.RECUPERAR_REDEFINIR && (
          <form onSubmit={aoRedefinirSenha} noValidate>
            <div className="campo-formulario">
              <label htmlFor="campo-resposta-recuperar">{perguntaExibida}</label>
              <input
                id="campo-resposta-recuperar"
                type="text"
                required
                value={respostaRecuperacao}
                onChange={(e) => setRespostaRecuperacao(e.target.value)}
              />
            </div>

            <div className="campo-formulario">
              <label htmlFor="campo-nova-senha">Nova senha</label>
              <input
                id="campo-nova-senha"
                type="password"
                autoComplete="new-password"
                required
                minLength={6}
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

        {/* ---------- Links para alternar entre os modos ---------- */}
        {modo === MODOS.ENTRAR && (
          <>
            <button
              type="button"
              className="alternar-modo"
              onClick={() => irParaModo(MODOS.CADASTRO)}
            >
              Ainda não tenho conta — criar
            </button>
            <button
              type="button"
              className="link-recuperar-senha"
              onClick={() => irParaModo(MODOS.RECUPERAR_PERGUNTA)}
            >
              Esqueci minha senha
            </button>
          </>
        )}

        {modo === MODOS.CADASTRO && (
          <button
            type="button"
            className="alternar-modo"
            onClick={() => irParaModo(MODOS.ENTRAR)}
          >
            Já tenho conta — entrar
          </button>
        )}

        {(modo === MODOS.RECUPERAR_PERGUNTA ||
          modo === MODOS.RECUPERAR_REDEFINIR) && (
          <button
            type="button"
            className="alternar-modo"
            onClick={() => irParaModo(MODOS.ENTRAR)}
          >
            Voltar para o login
          </button>
        )}
      </div>
    </main>
  );
}
