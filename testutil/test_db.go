package testutil

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest"
)

func CreateContainer() (*dockertest.Resource, *dockertest.Pool) {
	// Dockerコンテナへのファイルマウント時に絶対パスが必要
	pwd, _ := os.Getwd()

	// Dockerとの接続
	pool, err := dockertest.NewPool("")

	pool.MaxWait = time.Minute * 3
	if err != nil {
		log.Fatalf("Could not connect to docker ee: %s", err)
	}

	// Dockerコンテナ起動時の細かいオプションを指定する
	// テーブル定義などはここで流し込むのが良さそう
	runOptions := &dockertest.RunOptions{
		Repository: "mysql",
		Tag:        "8.0.29",
		Env: []string{
			"MYSQL_ROOT_PASSWORD=secret",
			"MYSQL_DATABASE=todo",
		},
		Mounts: []string{
			pwd + "/_tools/mysql/conf.d:/etc/mysql/conf.d",
		},
	}

	// コンテナを起動
	resource, err := pool.RunWithOptions(runOptions)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	return resource, pool
}

func CloseContainer(resource *dockertest.Resource, pool *dockertest.Pool) {
	// コンテナの終了
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

func ConnectDB(resource *dockertest.Resource, pool *dockertest.Pool) (*sqlx.DB, string) {
	// DB(コンテナ)との接続
	var db *sql.DB
	var port string
	if err := pool.Retry(func() error {
		// DBコンテナが立ち上がってから疎通可能になるまで少しかかるのでちょっと待ったほうが良さそう
		time.Sleep(time.Second * 50)

		var err error
		// port := 33306
		// if _, defind := os.LookupEnv("CI"); defind {
		// 	port = 3306
		// }
		port = resource.GetPort("3306/tcp")
		db, err = sql.Open("mysql", fmt.Sprintf("root:secret@(localhost:%s)/todo?parseTime=true", port))
		// log.Fatalf(fmt.Sprintf("root:secret@(localhost:%s)/mysql", resource.GetPort("3306/tcp")))
		// sql.Open("mysql", fmt.Sprintf("root:secret@tcp(%s)/todo?parseTime=true", resource.GetHostPort(fmt.Sprintf("%v/tcp", port))))
		if err != nil {
			log.Fatalf("Error opening database connection: %v", err)
			return err
		}
		query := "DROP TABLE IF EXISTS user;"
		_, err = db.Exec(query)
		if err != nil {
			fmt.Println("Failed to drop user table:", err)
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	return sqlx.NewDb(db, "mysql"), port
}
