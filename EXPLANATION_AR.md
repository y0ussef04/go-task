# 🧠 شرح النظام بالكامل — Dynamic Database Management System

## 📌 الفكرة العامة

النظام ده عبارة عن **نسخة مبسطة من phpMyAdmin** مكتوبة بـ Go.
بيخليك تعمل كل حاجة dynamically — يعني مفيش أي جدول أو عمود مكتوب في الكود ثابت.
كل حاجة بتتعمل وقت التشغيل (Runtime).

---

## 🏗️ هيكل المشروع (Project Structure)

```
go-task/
├── main.go                          ← نقطة البداية
├── internal/
│   ├── config/config.go             ← إعدادات الاتصال
│   ├── models/models.go             ← النماذج (Models)
│   ├── repository/mysql_repo.go     ← طبقة قاعدة البيانات
│   ├── service/db_service.go        ← طبقة منطق الأعمال
│   └── handler/handler.go           ← طبقة HTTP
├── static/
│   ├── index.html                   ← صفحة الويب
│   ├── styles.css                   ← التصميم
│   └── app.js                       ← الجافاسكريبت
├── Dockerfile                       ← بناء الصورة
└── docker-compose.yml               ← تشغيل الحاويات
```

---

## 🔄 فلو الداتا (Data Flow)

لما المستخدم يعمل أي عملية، الداتا بتمشي بالترتيب ده:

```
المتصفح (Browser)
    ↓  يبعت HTTP Request (مثلاً POST /api/databases)
    ↓
Handler (handler.go)
    ↓  يقرأ الـ JSON من الـ Request
    ↓  يعمل Validation أولي
    ↓
Service (db_service.go)
    ↓  يتحقق من البيانات (اسم فاضي؟ فيه أعمدة؟)
    ↓  ينادي على الـ Repository
    ↓
Repository (mysql_repo.go)
    ↓  يفتح اتصال بـ MySQL
    ↓  يبني الـ SQL Query
    ↓  ينفذ باستخدام Prepared Statements
    ↓
MySQL Database
    ↓  يرجّع النتيجة
    ↓
Repository → Service → Handler
    ↓  يحوّل النتيجة لـ JSON
    ↓
المتصفح (يعرض النتيجة)
```

---

## 📁 شرح كل ملف

---

### 1️⃣ `main.go` — نقطة البداية

```
الملف ده بيعمل 4 حاجات بس:
```

1. **يقرأ الإعدادات** من `config.Load()` — (اسم الهوست، الباسورد، البورت)
2. **يستنى MySQL** يكون جاهز — في loop بيحاول يتصل 30 مرة
3. **يربط الطبقات ببعض**: Repository → Service → Handler
4. **يشغّل السيرفر** على البورت 8081

```go
cfg := config.Load()                           // اقرأ الإعدادات
repo := repository.NewMySQLRepository(cfg.DSN)  // أنشئ الـ Repository
svc := service.NewDBService(repo)               // أنشئ الـ Service
h := handler.NewHandler(svc)                    // أنشئ الـ Handler
```

> ده اسمه **Dependency Injection** — كل طبقة بتاخد اللي تحتها كـ parameter.

---

### 2️⃣ `config/config.go` — الإعدادات

بيقرأ الإعدادات من **Environment Variables**:

| متغير | القيمة الافتراضية | الوصف |
|--------|---------|------|
| `DB_HOST` | `mysql` | اسم سيرفر MySQL (اسم الـ container في Docker) |
| `DB_PORT` | `3306` | بورت MySQL |
| `DB_USER` | `root` | اسم المستخدم |
| `DB_PASSWORD` | `secret` | الباسورد |
| `APP_PORT` | `8081` | البورت اللي التطبيق يشتغل عليه |

الدالة `DSN(dbName)` بتبني الـ Connection String:
```
root:secret@tcp(mysql:3306)/اسم_الداتابيز?parseTime=true
```

---

### 3️⃣ `models/models.go` — النماذج

فيه 3 أنواع أساسية:

```go
// عمود في جدول
type Column struct {
    Name string   // اسم العمود (مثلاً "Agent_name")
    Type string   // نوعه (مثلاً "VARCHAR(255)")
}

// صف في جدول — عبارة عن map من اسم العمود للقيمة
type Record map[string]interface{}

// الرد بتاع أي API
type APIResponse struct {
    Success bool        // نجح ولا لا
    Message string      // رسالة للمستخدم
    Error   string      // رسالة خطأ
    Data    interface{} // أي بيانات مرجعة
}
```

> `Record` هو `map[string]interface{}` — يعني ممكن يشيل أي عمود بأي قيمة. ده اللي بيخلي النظام **dynamic**.

---

### 4️⃣ `repository/mysql_repo.go` — طبقة قاعدة البيانات ⭐

**دي أهم طبقة** — هي اللي بتكلم MySQL مباشرة.

#### كيف بيتصل بالداتابيز؟

```go
func (r *MySQLRepository) open(dbName string) (*sql.DB, error) {
    db, err := sql.Open("mysql", r.baseDSN(dbName))
    // ...
}
```

كل عملية بتفتح اتصال جديد بالداتابيز المطلوبة — مش بيستخدم اتصال واحد ثابت.
ده لأن المستخدم ممكن يتعامل مع أكثر من داتابيز.

#### العمليات الموجودة:

| الدالة | بتعمل إيه |
|--------|----------|
| `CreateDatabase(name)` | `CREATE DATABASE IF NOT EXISTS` |
| `ListDatabases()` | `SHOW DATABASES` |
| `ListTables(dbName)` | `SHOW TABLES` |
| `CreateTable(dbName, tableName, columns)` | `CREATE TABLE IF NOT EXISTS` |
| `GetColumns(dbName, tableName)` | `SHOW COLUMNS FROM` — بترجع اسم ونوع كل عمود |
| `AddColumn(dbName, tableName, colName, colType)` | `ALTER TABLE ... ADD COLUMN` |
| `InsertRecord(dbName, tableName, data)` | `INSERT INTO ... VALUES (?, ?, ?)` |
| `SelectRecords(dbName, tableName)` | `SELECT * FROM` — بترجع كل الصفوف |
| `UpdateRecord(dbName, tableName, data, condCol, condVal)` | `UPDATE ... SET ... WHERE col = ?` |
| `DeleteRecord(dbName, tableName, condCol, condVal)` | `DELETE FROM ... WHERE col = ?` |

#### 🔒 الأمان:

1. **Identifier Validation** — كل اسم (جدول/عمود/داتابيز) لازم يكون حروف وأرقام و `_` بس:
```go
func isValidIdentifier(name string) bool {
    for _, c := range name {
        if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || 
             (c >= '0' && c <= '9') || c == '_') {
            return false
        }
    }
    return true
}
```

2. **Backtick Quoting** — كل الأسماء بتتحط في `` ` `` في الـ SQL:
```sql
INSERT INTO `Agent` (`Agent_name`, `Agent_salary`) VALUES (?, ?)
```

3. **Prepared Statements** — القيم بتتبعت كـ `?` مش بتتحط في النص:
```go
stmt, err := db.Prepare(query)    // حضّر الـ Query
_, err = stmt.Exec(values...)     // نفّذ بالقيم
```

> ده بيحمي من **SQL Injection**. لو حد كتب `'; DROP TABLE--` كقيمة، MySQL هيتعامل معاها كـ text عادي.

---

### 5️⃣ `service/db_service.go` — طبقة منطق الأعمال

الطبقة دي **رفيعة** (thin layer) — وظيفتها:

1. **التحقق من المدخلات** — مثلاً: الاسم فاضي؟ مفيش أعمدة؟
2. **تنادي على الـ Repository** — تمرر البيانات بعد التحقق
3. **تجمع عمليات معقدة** — مثلاً `CreateSampleDatabase()` بتعمل:
   - إنشاء داتابيز `RealEstate`
   - إنشاء 3 جداول (Campaign, Agent, Properties)
   - كل ده في دالة واحدة

```go
func (s *DBService) InsertRecord(dbName, tableName string, data map[string]interface{}) error {
    if dbName == "" || tableName == "" {
        return fmt.Errorf("database and table names are required")  // تحقق
    }
    return s.repo.InsertRecord(dbName, tableName, data)  // مرّر للـ Repository
}
```

---

### 6️⃣ `handler/handler.go` — طبقة HTTP

بتستقبل الـ Requests من المتصفح وبتحولها لنداءات على الـ Service.

#### الراوتات (Routes):

```
GET    /api/databases                           → قائمة الداتابيزات
POST   /api/databases                           → إنشاء داتابيز
GET    /api/databases/{db}/tables               → قائمة الجداول
POST   /api/databases/{db}/tables               → إنشاء جدول
GET    /api/databases/{db}/tables/{t}/columns   → أعمدة الجدول
POST   /api/databases/{db}/tables/{t}/columns   → إضافة عمود
GET    /api/databases/{db}/tables/{t}/records   → عرض السجلات
POST   /api/databases/{db}/tables/{t}/records   → إدخال سجل
PUT    /api/databases/{db}/tables/{t}/records   → تعديل سجل
DELETE /api/databases/{db}/tables/{t}/records   → حذف سجل
POST   /api/sample                              → إنشاء داتابيز تجريبي
```

#### كيف بيشتغل الـ Router:

الـ `handleDatabaseSub` بياخد الـ URL ويقطّعه:

```
/api/databases/RealEstate/tables/Agent/records
         ↓          ↓        ↓       ↓
      parts[0]   parts[1]  parts[2] parts[3]
     "RealEstate" "tables"  "Agent"  "records"
```

وبعدين بيوجّه لداله المناسبة حسب `parts[3]` و HTTP Method.

---

### 7️⃣ `static/app.js` — الفرونت إند

الجافاسكريبت بيتعامل مع الـ API عن طريق `fetch()`.

#### مثال: إدخال سجل

```
1. المستخدم يختار جدول من الـ dropdown
2. الكود يبعت GET /api/databases/{db}/tables/{t}/schema
3. MySQL يرجّع الأعمدة (اسم + نوع)
4. الكود يبني form تلقائياً بناءً على الأعمدة
5. المستخدم يملا البيانات ويضغط Save
6. الكود يبعت POST /api/databases/{db}/tables/{t}/records
7. السيرفر يعمل INSERT INTO ... VALUES (?, ?)
```

> **ده معنى "Dynamic"** — الفورم مش مكتوبة في الـ HTML. بتتولد من schema الجدول الفعلي.

#### مثال: تعديل سجل

```
1. المستخدم يضغط ✏️ على صف
2. يفتح Modal بالبيانات الحالية
3. يعدّل ويضغط Save
4. الكود يبعت PUT مع:
   - data: القيم الجديدة
   - condition_col: اسم العمود المفتاح (أول عمود)
   - condition_val: قيمة المفتاح
5. السيرفر ينفذ: UPDATE `Agent` SET `Agent_name` = ? WHERE `Agent_id` = ?
```

---

## 🐳 Docker

### Dockerfile — بناء التطبيق

```dockerfile
# المرحلة 1: بناء الكود
FROM golang:1.22-alpine AS builder
# ... يعمل compile للـ Go code

# المرحلة 2: تشغيل
FROM alpine:3.19
# ... ينسخ الـ binary بس (بدون Go compiler)
```

> **Multi-stage Build** = الصورة النهائية صغيرة جداً (< 20MB) بدلاً من 800MB+

### docker-compose.yml — التنسيق

```yaml
services:
  mysql:    # حاوية MySQL
  app:      # حاوية التطبيق (بتستنى MySQL يكون جاهز)
```

**الـ Healthcheck** بيتحقق إن MySQL شغال قبل ما التطبيق يبدأ:
```yaml
healthcheck:
  test: ["CMD", "mysqladmin", "ping", "-h", "localhost", "-psecret"]
```

التطبيق كمان عنده retry loop خاص بيه كطبقة حماية إضافية.

---

## 🔑 مفاهيم مهمة

### Clean Architecture (العمارة النظيفة)
```
Handler  →  لا يعرف أي شيء عن SQL
Service  →  لا يعرف أي شيء عن HTTP
Repository → لا يعرف أي شيء عن الـ Web
```
كل طبقة مسؤولة عن شيء واحد فقط. ده بيخلي الكود:
- ✅ سهل الفهم
- ✅ سهل التعديل
- ✅ سهل الاختبار

### Prepared Statements
بدلاً من:
```sql
-- ❌ خطير (SQL Injection)
INSERT INTO Agent (name) VALUES ('يوسف')
```
بنستخدم:
```sql
-- ✅ آمن
INSERT INTO Agent (name) VALUES (?)
-- والقيمة 'يوسف' بتتبعت منفصلة
```

### Dynamic Schema
النظام مش عارف أي جدول هيتعمل أو أي أعمدة فيه.
بيقرأ الـ Schema من MySQL نفسه باستخدام `SHOW COLUMNS FROM` وبيبني كل حاجة وقت التشغيل.

---

## ⚡ التشغيل

```bash
docker-compose up --build
```

بعدها افتح: **http://localhost:8081**

1. اضغط **🎲 Sample Database** لإنشاء داتابيز تجريبي
2. اضغط على **RealEstate** في الـ Dashboard
3. روح **Insert Record** واختار جدول
4. املا البيانات واحفظ
5. روح **Browse Data** وشوف البيانات واعدّل أو امسح
