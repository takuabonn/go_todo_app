package store

import (
	"context"
	"os"
	"os/exec"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-cmp/cmp"
	"github.com/takuabonn/go_todo_app/clock"
	"github.com/takuabonn/go_todo_app/entity"
	"github.com/takuabonn/go_todo_app/testutil"
)

func runMigration(port string) error {

	filePath := "../_tools/mysql/schema.sql"
	cmd := exec.Command("mysqldef",
		"-u", "root",
		"-p", "secret",
		"-h", "localhost",
		"-P", port,
		"todo",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// ファイルからの入力を設定
	inFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer inFile.Close()
	cmd.Stdin = inFile

	return cmd.Run()
}

func prepareTasks(ctx context.Context, t *testing.T, con Execer) entity.Tasks {
	t.Helper()
	c := clock.FixedClocker{}
	wants := entity.Tasks{
		{
			Title: "want task 1", Status: "todo",
			Created: c.Now(), Modified: c.Now(),
		},
		{
			Title: "want task 2", Status: "todo",
			Created: c.Now(), Modified: c.Now(),
		},
		{
			Title: "want task 3", Status: "done",
			Created: c.Now(), Modified: c.Now(),
		},
	}
	result, err := con.ExecContext(ctx,
		`INSERT INTO task (title, status, created, modified)
			VALUES
			    (?, ?, ?, ?),
			    (?, ?, ?, ?),
			    (?, ?, ?, ?);`,
		wants[0].Title, wants[0].Status, wants[0].Created, wants[0].Modified,
		wants[1].Title, wants[1].Status, wants[1].Created, wants[1].Modified,
		wants[2].Title, wants[2].Status, wants[2].Created, wants[2].Modified,
	)
	if err != nil {
		t.Fatal(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	wants[0].ID = entity.TaskID(id)
	wants[1].ID = entity.TaskID(id + 1)
	wants[2].ID = entity.TaskID(id + 2)
	return wants
}

func TestRepository_ListTasks(t *testing.T) {
	ctx := context.Background()
	// コンテナ(DB)の立ち上げ, 接続
	resource, pool := testutil.CreateContainer()

	defer testutil.CloseContainer(resource, pool)
	db, port := testutil.ConnectDB(resource, pool)

	err := runMigration(port)
	if err != nil {
		t.Fatalf("Migration failed: %s", err)
	}
	wants := prepareTasks(ctx, t, db)

	repository := &Repository{}

	gots, err := repository.ListTasks(ctx, db)
	t.Log(gots)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d := cmp.Diff(gots, wants); len(d) != 0 {
		t.Errorf("differs: (-got +want)\n%s", d)
	}

}

func TestRepository_AddTask(t *testing.T) {
	ctx := context.Background()
	resource, pool := testutil.CreateContainer()

	defer testutil.CloseContainer(resource, pool)
	db, port := testutil.ConnectDB(resource, pool)

	err := runMigration(port)
	if err != nil {
		t.Fatalf("Migration failed: %s", err)
	}
	c := clock.FixedClocker{}
	repository := &Repository{Clocker: c}
	insertTask := &entity.Task{
		Title:    "want task 1",
		Status:   "todo",
		Created:  c.Now(),
		Modified: c.Now(),
	}

	if err := repository.AddTask(ctx, db, insertTask); err != nil {
		t.Fatalf("cannot insert data: %s", err)
	}

	tasks := &entity.Tasks{}
	sql := `SELECT
			id, title,
			status, created, modified
		FROM task where id = ? limit 1;`
	if err := db.SelectContext(ctx, tasks, sql, insertTask.ID); err != nil {
		t.Errorf("not found task :%s", err)
	}

}
