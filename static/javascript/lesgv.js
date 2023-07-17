function addResDigitaOrgIdValue () {
  inputTextEl = document.getElementById("idvalue");
  value = inputTextEl.value;
  value = value.split("@")[0];
  inputTextEl.value = value + "@resdigita.org";
  return value;
}
document.getElementById("idvalue").addEventListener("change",addResDigitaOrgIdValue());