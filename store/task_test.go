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
	"github.com/takuabonn/go_todo_app/testutil/fixture"
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

func prepareUser(ctx context.Context, t *testing.T, db Execer) entity.UserID {
	t.Helper()
	u := fixture.User(nil)
	result, err := db.ExecContext(ctx,
		`INSERT INTO user (name, password, role, created, modified)
		VALUES (?, ?, ?, ?, ?);`,
		u.Name, u.Password, u.Role, u.Created, u.Modified,
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("got user_id: %v", err)
	}

	return entity.UserID(id)
}
func prepareTasks(ctx context.Context, t *testing.T, con Execer) (entity.UserID, entity.Tasks) {
	t.Helper()
	userID := prepareUser(ctx, t, con)
	otherUserID := prepareUser(ctx, t, con)
	c := clock.FixedClocker{}
	wants := entity.Tasks{
		{
			UserID: userID,
			Title:  "want task 1", Status: "todo",
			Created: c.Now(), Modified: c.Now(),
		},
		{
			UserID: userID,
			Title:  "want task 2", Status: "todo",
			Created: c.Now(), Modified: c.Now(),
		},
	}
	tasks := entity.Tasks{
		wants[0],
		{
			UserID:  otherUserID,
			Title:   "not want task",
			Status:  "todo",
			Created: c.Now(), Modified: c.Now(),
		},
		wants[1],
	}
	result, err := con.ExecContext(ctx,
		`INSERT INTO task (user_id, title, status, created, modified)
			VALUES
			    (?, ?, ?, ?, ?),
			    (?, ?, ?, ?, ?),
			    (?, ?, ?, ?, ?);`,
		tasks[0].UserID, tasks[0].Title, tasks[0].Status, tasks[0].Created, tasks[0].Modified,
		tasks[1].UserID, tasks[1].Title, tasks[1].Status, tasks[1].Created, tasks[1].Modified,
		tasks[2].UserID, tasks[2].Title, tasks[2].Status, tasks[2].Created, tasks[2].Modified,
	)
	if err != nil {
		t.Fatal(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	tasks[0].ID = entity.TaskID(id)
	tasks[1].ID = entity.TaskID(id + 1)
	tasks[2].ID = entity.TaskID(id + 2)
	return userID, wants
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
	wantUserID, wants := prepareTasks(ctx, t, db)

	repository := &Repository{}

	gots, err := repository.ListTasks(ctx, db, wantUserID)
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
	userID := prepareUser(ctx, t, db)
	insertTask := &entity.Task{
		UserID:   userID,
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
			id, user_id, title,
			status, created, modified
		FROM task where id = ? limit 1;`
	if err := db.SelectContext(ctx, tasks, sql, insertTask.ID); err != nil {
		t.Errorf("not found task :%s", err)
	}
}
