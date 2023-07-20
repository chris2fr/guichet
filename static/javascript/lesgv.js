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


function changeUsername () {
   username = document.getElementById("username");
   calcCn = document.getElementById("calc-cn");
  calcCn.innerText = "Login Name et Courriel seront : " + username.value.split("@")[0] + "@lesgv.com";
}
if (document.getElementById("username") != null) {
  document.getElementById("username").addEventListener("change",changeUsername);
}
