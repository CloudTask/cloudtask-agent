package notify

var OKSubjectPrefix = "(info) "
var FAILSubjectPrefix = "(fail) "
var WARNSubjectPrefix = "(warn) "

var MailTemplate = `
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta charset="utf-8">
  <meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
  <title>Cloud Task</title>
  <style type="text/css">
    td.title {
      height: 25px;
      color: #FFFFFF;
      border-bottom: solid 1px #89d7f1;
      width: 100px;
      background-color: #2a7b97;
      text-align: left;
      padding-left: 10px;
    }
    
    td.content {
      padding-left: 10px;
      padding-right: 20px;
      height: 22px;
      border-bottom: solid 1px #89d7f1;
      background-color: #7ecae4;
      word-wrap: break-word;
      word-break: break-all;
    }
    
    td.output> div {
      background-color: #7ecae4;
    }
  </style>
</head>
<body>
  <table width="800" border="0" align="center" cellpadding="0" cellspacing="0" bgcolor="#b1ebff" style="padding: 0 10px; font-family: Verdana, Arial, Helvetica, sans-serif; font-size: 12px; color: #333;">
    <tr>
      <td colspan="2" style="padding:0 10px;">
        <a href="task.newegg.org" target="_blank" style=" color: #17627c; text-decoration: none;">
          <div>
            <strong style="font-size: 40px; color: white; text-shadow: 3px 3px 3px gray; filter: dropshadow(color=#000000, offx=2, offy=2);">Cloud</strong>
            <strong style="font-size: 60px; text-shadow: 3px 3px 3px gray; filter: dropshadow(color=#000000, offx=2, offy=2);">Task</strong>
          </div>
        </a>
      </td>
    </tr>
    <tr>
      <td colspan="2" style="border-bottom: solid 1px #89d7f1;"></td>
    </tr>
    <tr>
      <td colspan="2">&nbsp;</td>
    </tr>
    <tr>
      <td class="title"><strong>Job name</strong></td>
      <td class="content"><strong>{{.jobName}}</strong></td>
    </tr>
    <tr>
      <td class="title"><strong>Location</strong></td>
      <td class="content"><strong>{{.location}}</strong></td>
    </tr>
    <tr>
      <td class="title"><strong>Server</strong></td>
      <td class="content"><strong>{{.server}}</strong></td>
    </tr>
    <tr>
      <td class="title"><strong>Start time</strong></td>
      <td class="content"><strong>{{.execat}}</strong></td>
    </tr>
    <tr>
      <td class="title"><strong>Duration</strong></td>
      <td class="content"><strong>{{.duration}}</strong></td>
    </tr>
    <tr>
      <td class="title"><strong>Result</strong></td>
      <td class="content">
        {{ if eq .isSucceed true }}
        <strong style="color: green">Succeed</strong> {{ else }}
        <strong style="color: red">Failed</strong> {{ end }}
      </td>
    </tr>
    <tr>
      <td class="title"><strong>Directory</strong></td>
      <td class="content"><strong>{{.directory}}</strong></td>
    </tr>
    <tr>
      <td class="title"><strong>Content</strong></td>
      <td class="content"><strong>{{.content}}</strong></td>
    </tr>
    <tr>
      <td class="title" style="vertical-align: top;">
        <div style="height: 25px; line-height: 25px; font-weight: 700;">Output</div>
      </td>
      <td class="content output">
        {{ if .stdout }}
        <div style="font-weight: 600; height: 25px; line-height: 25px;">Stdout</div>
        <pre>{{.stdout}}</pre>
        {{ end }} {{ if .errout }}
        <div style="font-weight: 600; height: 25px; line-height: 25px; margin-top: 10px;">Errout</div>
        <pre>{{.errout}}</pre>
        {{ end }} {{ if .execerr }}
        <div style="font-weight: 600; height: 25px; line-height: 25px; margin-top: 10px;">Exec Err</div>
        <pre>{{.execerr}}</pre>
        {{ end }}
      </td>
    </tr>
    <tr>
      <td colspan="2" style="padding-left: 10px; color: #FFFFFF; height:22px; ">&nbsp;</td>
    </tr>
  </table>
</body>
</html>
`
