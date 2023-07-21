function addResDigitaOrg (id) {
  if (document.getElementById(id).value.split("@")[1] && !["lesgrandsvoisins.com","resdigita.com","resdigita.org","lesgv.com","lesgv.org"].includes(document.getElementById(id).value.split("@")[1])) {
    document.getElementById(id).value = document.getElementById(id).value.split("@")[0] + "@lesgrandsvoisins.com";
  }
  return document.getElementById(id).value;
}
function addResDigitaOrgIdValue () {
  document.getElementById("mail").value = addResDigitaOrg("idvalue");
}
function addResDigitaOrgMail () {
  document.getElementById("idvalue").value = addResDigitaOrg("mail");
}
let idvalueInput = document.querySelector("#idvalue");
if (idvalueInput != null) {
  idvalueInput.addEventListener("change",addResDigitaOrgIdValue);
}



