// Code generated by qtc from "banned.qtpl". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

//line templates/banned.qtpl:1
package templates

//line templates/banned.qtpl:1
import "github.com/catatsuy/private-isu/webapp/golang/types"

//line templates/banned.qtpl:3
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line templates/banned.qtpl:3
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line templates/banned.qtpl:3
func StreamAdminBannedPage(qw422016 *qt422016.Writer, users []types.User, csrfToken string) {
//line templates/banned.qtpl:3
	qw422016.N().S(`
<div>
  <form method="post" action="/admin/banned">
    `)
//line templates/banned.qtpl:6
	for _, user := range users {
//line templates/banned.qtpl:6
		qw422016.N().S(`
    <div>
      <input type="checkbox" name="uid[]" id="uid_`)
//line templates/banned.qtpl:8
		qw422016.N().D(user.ID)
//line templates/banned.qtpl:8
		qw422016.N().S(`" value="`)
//line templates/banned.qtpl:8
		qw422016.N().D(user.ID)
//line templates/banned.qtpl:8
		qw422016.N().S(`" data-account-name="`)
//line templates/banned.qtpl:8
		qw422016.E().S(user.AccountName)
//line templates/banned.qtpl:8
		qw422016.N().S(`"> <label for="uid_`)
//line templates/banned.qtpl:8
		qw422016.N().D(user.ID)
//line templates/banned.qtpl:8
		qw422016.N().S(`">`)
//line templates/banned.qtpl:8
		qw422016.E().S(user.AccountName)
//line templates/banned.qtpl:8
		qw422016.N().S(`</label>
    </div>
    `)
//line templates/banned.qtpl:10
	}
//line templates/banned.qtpl:10
	qw422016.N().S(`
    <div class="form-submit">
      <input type="hidden" name="csrf_token" value="`)
//line templates/banned.qtpl:12
	qw422016.E().S(csrfToken)
//line templates/banned.qtpl:12
	qw422016.N().S(`">
      <input type="submit" name="submit" value="submit">
    </div>
  </form>
</div>
`)
//line templates/banned.qtpl:17
}

//line templates/banned.qtpl:17
func WriteAdminBannedPage(qq422016 qtio422016.Writer, users []types.User, csrfToken string) {
//line templates/banned.qtpl:17
	qw422016 := qt422016.AcquireWriter(qq422016)
//line templates/banned.qtpl:17
	StreamAdminBannedPage(qw422016, users, csrfToken)
//line templates/banned.qtpl:17
	qt422016.ReleaseWriter(qw422016)
//line templates/banned.qtpl:17
}

//line templates/banned.qtpl:17
func AdminBannedPage(users []types.User, csrfToken string) string {
//line templates/banned.qtpl:17
	qb422016 := qt422016.AcquireByteBuffer()
//line templates/banned.qtpl:17
	WriteAdminBannedPage(qb422016, users, csrfToken)
//line templates/banned.qtpl:17
	qs422016 := string(qb422016.B)
//line templates/banned.qtpl:17
	qt422016.ReleaseByteBuffer(qb422016)
//line templates/banned.qtpl:17
	return qs422016
//line templates/banned.qtpl:17
}