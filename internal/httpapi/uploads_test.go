package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/GoreeCloud/goreecloud-photos/internal/upload"
)

type subjectResolverFunc func(*http.Request) (string, error)
func (f subjectResolverFunc) ResolveSubject(r *http.Request) (string, error) { return f(r) }

type uploadServiceStub struct {
	create func(context.Context,string,upload.CreateRequest)(upload.Session,error)
	get func(context.Context,string,string)(upload.Session,error)
	put func(context.Context,string,string,int,io.Reader)(upload.Session,error)
	cancel func(context.Context,string,string)(upload.Session,error)
}
func (s uploadServiceStub) Create(c context.Context,a string,r upload.CreateRequest)(upload.Session,error){return s.create(c,a,r)}
func (s uploadServiceStub) Get(c context.Context,a,u string)(upload.Session,error){return s.get(c,a,u)}
func (s uploadServiceStub) PutPart(c context.Context,a,u string,n int,r io.Reader)(upload.Session,error){return s.put(c,a,u,n,r)}
func (s uploadServiceStub) Cancel(c context.Context,a,u string)(upload.Session,error){return s.cancel(c,a,u)}

func TestUploadRoutesFailClosedWithoutIdentity(t *testing.T) {
	server:=New("0.1.0-experimental.2","experimental",Dependencies{Uploads:uploadServiceStub{}})
	req:=httptest.NewRequest(http.MethodPost,"/api/v1/libraries/01999999-0000-7000-8000-000000000001/uploads",strings.NewReader(`{"original_filename":"photo.jpg","media_type":"image/jpeg","expected_size":4}`))
	res:=httptest.NewRecorder(); server.ServeHTTP(res,req)
	if res.Code!=http.StatusServiceUnavailable { t.Fatalf("status=%d",res.Code) }
	var body apiErrorResponse
	if err:=json.Unmarshal(res.Body.Bytes(),&body);err!=nil{t.Fatal(err)}
	if body.Code!="identity_unavailable"||body.RequestID==""{t.Fatalf("body=%#v",body)}
	if res.Header().Get("X-Request-ID")==""{t.Fatal("missing request id")}
}

func TestCreateUploadUsesResolvedSubject(t *testing.T) {
	const libraryID="01999999-0000-7000-8000-000000000001"
	expected:=upload.Session{UploadID:"01999999-0000-7000-8000-000000000002",LibraryID:libraryID,ExpectedSize:4,PartSize:upload.DefaultPartSize,State:"open",TransferMethod:"parts-v1",ExpiresAt:time.Date(2026,9,19,0,0,0,0,time.UTC),CreatedAt:time.Date(2026,9,18,0,0,0,0,time.UTC),UpdatedAt:time.Date(2026,9,18,0,0,0,0,time.UTC),Parts:[]upload.Part{}}
	server:=New("0.1.0-experimental.2","experimental",Dependencies{
		Identity:subjectResolverFunc(func(*http.Request)(string,error){return "subject-1",nil}),
		Uploads:uploadServiceStub{create:func(_ context.Context,actor string,r upload.CreateRequest)(upload.Session,error){
			if actor!="subject-1"||r.LibraryID!=libraryID||r.OriginalFilename!="photo.jpg"{t.Fatalf("actor=%q request=%#v",actor,r)}
			return expected,nil
		}},
	})
	req:=httptest.NewRequest(http.MethodPost,"/api/v1/libraries/"+libraryID+"/uploads",strings.NewReader(`{"original_filename":"photo.jpg","media_type":"image/jpeg","expected_size":4}`))
	res:=httptest.NewRecorder(); server.ServeHTTP(res,req)
	if res.Code!=http.StatusCreated{t.Fatalf("status=%d body=%s",res.Code,res.Body.String())}
	var actual upload.Session
	if err:=json.Unmarshal(res.Body.Bytes(),&actual);err!=nil{t.Fatal(err)}
	if actual.UploadID!=expected.UploadID{t.Fatalf("session=%#v",actual)}
}

func TestUploadErrorDoesNotExposeInternalError(t *testing.T) {
	server:=New("0.1.0-experimental.2","experimental",Dependencies{
		Identity:subjectResolverFunc(func(*http.Request)(string,error){return "subject-1",nil}),
		Uploads:uploadServiceStub{get:func(context.Context,string,string)(upload.Session,error){return upload.Session{},errors.New("postgres://secret:password@database")}},
	})
	req:=httptest.NewRequest(http.MethodGet,"/api/v1/uploads/01999999-0000-7000-8000-000000000002",nil)
	res:=httptest.NewRecorder(); server.ServeHTTP(res,req)
	if res.Code!=http.StatusInternalServerError{t.Fatalf("status=%d",res.Code)}
	if strings.Contains(res.Body.String(),"password")||strings.Contains(res.Body.String(),"postgres://"){t.Fatalf("leak: %s",res.Body.String())}
}
