function addResDigitaOrg (id) {
  document.getElementById(id).value = document.getElementById(id).value.split("@")[0] + "@resdigita.org";
  return document.getElementById(id).value;
}
function addResDigitaOrgIdValue () {
  document.getElementById("mail").value = addResDigitaOrg("idvalue");
}
function addResDigitaOrgMail () {
  document.getElementById("idvalue").value = addResDigitaOrg("mail");
}
document.getElementById("mail").addEventListener("change",addResDigitaOrgMail);
document.getElementById("idvalue").addEventListener("change",addResDigitaOrgIdValue);