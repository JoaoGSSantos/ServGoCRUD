package main

import (
	"database/sql" // Pacote Database SQL para realizar Query
	"log"
	"net/http" // Gerencia URLs e Servidor Web
	"os"
	"strconv"
	"text/template" // Gerencia templates
	"time"

	_ "github.com/go-sql-driver/mysql" // Driver Mysql para Go
)

//Struct utilizada para exibir dados no template
type Compromisso struct {
	Id        int
	Descricao string
	DataHora  time.Time
}

type Compromissos []Compromisso

func (c Compromissos) ShowHname() string {
	hname, err := os.Hostname()

	if err != nil {
		panic(err.Error())
	}

	return hname
}

//Renderiza todos os templates da pasta "tmpl" independente da extensão
var tmpl = template.Must(template.ParseGlob("tmpl/*"))

// Função dbConn, abre a conexão com o banco de dados
func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := "root"
	dbName := "agenda"

	dbEndpoint := "/" //jgss - Adicionado para configurar com RDS posteriormente

	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@"+dbEndpoint+dbName+"?parseTime=true")
	if err != nil {
		panic(err.Error())
	}
	return db
}

// Função usada para renderizar o arquivo Index
func Index(w http.ResponseWriter, r *http.Request) {
	// Abre a conexão com o banco de dados utilizando a função dbConn()
	db := dbConn()
	// Realiza a consulta com banco de dados e trata erros
	selDB, err := db.Query("SELECT * FROM compromissos ORDER BY id DESC")
	if err != nil {
		panic(err.Error())
	}

	// Monta a struct para ser utilizada no template
	n := Compromisso{}

	// Monta um array para guardar os valores da struct
	res := Compromissos{}

	// Realiza a estrutura de repetição pegando todos os valores do banco
	for selDB.Next() {
		// Armazena os valores em variáveis
		var id int
		var descricao string
		var data_hora time.Time

		// Faz o Scan do SELECT
		err = selDB.Scan(&id, &descricao, &data_hora)
		if err != nil {
			panic(err.Error())
		}

		// Envia os resultados para a struct
		n.Id = id
		n.Descricao = descricao
		n.DataHora = data_hora

		// Junta a Struct com Array
		res = append(res, n)
	}

	// Abre a página Index e exibe todos os registrados na tela
	tmpl.ExecuteTemplate(w, "Index", res)

	// Fecha a conexão
	defer db.Close()
}

// Função Show exibe apenas um resultado
func Show(w http.ResponseWriter, r *http.Request) {
	db := dbConn()

	// Pega o ID do parametro da URL
	nId := r.URL.Query().Get("id")

	// Usa o ID para fazer a consulta e tratar erros
	selDB, err := db.Query("SELECT * FROM compromissos WHERE id=?", nId)
	if err != nil {
		panic(err.Error())
	}

	// Monta a strcut para ser utilizada no template
	n := Compromisso{}

	// Realiza a estrutura de repetição pegando todos os valores do banco
	for selDB.Next() {
		// Armazena os valores em variaveis
		var id int
		var descricao string
		var data_hora time.Time

		// Faz o Scan do SELECT
		err = selDB.Scan(&id, &descricao, &data_hora)
		if err != nil {
			panic(err.Error())
		}

		// Envia os resultados para a struct
		n.Id = id
		n.Descricao = descricao
		n.DataHora = data_hora
	}

	// Mostra o template
	tmpl.ExecuteTemplate(w, "Show", n)

	// Fecha a conexão
	defer db.Close()

}

// Função New apenas exibe o formulário para inserir novos dados
func New(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "New", nil)
}

// Função Edit, edita os dados
func Edit(w http.ResponseWriter, r *http.Request) {
	// Abre a conexão com banco de dados
	db := dbConn()

	// Pega o ID do parametro da URL
	nId := r.URL.Query().Get("id")

	selDB, err := db.Query("SELECT * FROM compromissos WHERE id=?", nId)
	if err != nil {
		panic(err.Error())
	}

	// Monta a struct para ser utilizada no template
	n := Compromisso{}

	// Realiza a estrutura de repetição pegando todos os valores do banco
	for selDB.Next() {
		//Armazena os valores em variaveis
		var id int
		var descricao string
		var data_hora time.Time

		// Faz o Scan do SELECT
		err = selDB.Scan(&id, &descricao, &data_hora)
		if err != nil {
			panic(err.Error())
		}

		// Envia os resultados para a struct
		n.Id = id
		n.Descricao = descricao
		n.DataHora = data_hora
	}

	// Mostra o template com formulário preenchido para edição
	tmpl.ExecuteTemplate(w, "Edit", n)

	// Fecha a conexão com o banco de dados
	defer db.Close()
}

// Função Insert, insere valores no banco de dados
func Insert(w http.ResponseWriter, r *http.Request) {

	//Abre a conexão com banco de dados usando a função: dbConn()
	db := dbConn()

	// Verifica o METHOD do fomrulário passado
	if r.Method == "POST" {

		// Pega os campos do formulário
		descricao := r.FormValue("descricao")
		data_hora := r.FormValue("data_hora")

		// Prepara a SQL e verifica errors
		insForm, err := db.Prepare("INSERT INTO compromissos(descricao, data_hora) VALUES(?,?)")
		if err != nil {
			panic(err.Error())
		}

		// Insere valores do formulario com a SQL tratada e verifica errors
		insForm.Exec(descricao, data_hora)

		// Exibe um log com os valores digitados no formulário
		log.Println("INSERT: Descricao: " + descricao + " | Data/Hora: " + data_hora)
	}

	// Encerra a conexão do dbConn()
	defer db.Close()

	//Retorna a HOME
	http.Redirect(w, r, "/", http.StatusMovedPermanently) // Era 301
}

// Função Update, atualiza valores no banco de dados
func Update(w http.ResponseWriter, r *http.Request) {

	// Abre a conexão com o banco de dados usando a função: dbConn()
	db := dbConn()

	// Verifica o METHOD do formulário passado
	if r.Method == "POST" {

		// Pega os campos do formulário
		descricao := r.FormValue("descricao")
		data_hora, err := time.Parse("2006-01-02 15:04:05 -0700 MST", r.FormValue("data_hora"))

		if err != nil {
			panic(err.Error())
		}

		id, err := strconv.Atoi(r.FormValue("uid"))

		if err != nil {
			panic(err.Error())
		}

		// Prepara a SQL e verifica errors
		insForm, err := db.Prepare("UPDATE compromissos SET descricao=?, data_hora=? WHERE id=?")
		if err != nil {
			panic(err.Error())
		}

		// Insere valores do formulário com a SQL tratada e verifica erros
		insForm.Exec(descricao, data_hora, id)

		// Exibe um log com os valores digitados no formulario
		log.Println("UPDATE: Descricao: " + descricao)
		log.Println("|Data/Hora: " + data_hora.String())
		log.Printf("| Id: %d\n", id)
	}

	// Encerra a conexão do dbConn()
	defer db.Close()

	// Retorna a HOME
	http.Redirect(w, r, "/", http.StatusMovedPermanently) // Era 301
}

// Função Delete, deleta valores no banco de dados
func Delete(w http.ResponseWriter, r *http.Request) {

	// Abre conexão com banco de dados usando a função: dbConn()
	db := dbConn()

	nId := r.URL.Query().Get("id")

	// Prepara a SQL e verifica errors
	delForm, err := db.Prepare("DELETE FROM compromissos WHERE id=?")
	if err != nil {
		panic(err.Error())
	}

	// Insere valores do form com a SQL tratada e verifica errors
	delForm.Exec(nId)

	// Exibe um log com os valores digitados no form
	log.Println("DELETE")

	// Encerra a conexão do dbConn()
	defer db.Close()

	// Retorna a HOME
	http.Redirect(w, r, "/", http.StatusMovedPermanently) // Era 301
}

func main() {

	// Exibe mensagem que o servidor foi iniciado
	log.Println("Server started on: http://localhost:9000")
	log.Println(os.Hostname())

	// Gerencia as URLs
	http.HandleFunc("/", Index)
	http.HandleFunc("/show", Show)
	http.HandleFunc("/new", New)
	http.HandleFunc("/edit", Edit)

	// Ações
	http.HandleFunc("/insert", Insert)
	http.HandleFunc("/update", Update)
	http.HandleFunc("/delete", Delete)

	// Inicia o servidor na porta 9000
	http.ListenAndServe(":9000", nil)
}
