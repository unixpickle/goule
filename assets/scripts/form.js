function postData(fieldName, fieldValue, optionalTarget) {
    var $form = $('<form method="POST"><input name="' + fieldName + '" type="hidden"></form>');
    $form.find('input').val(taskJSON);
    if (optionalTarget) {
        $form.prop('target', optionalTarget);
    }
    $(document.body).append($form);
    $form.submit();
    $form.remove();
}