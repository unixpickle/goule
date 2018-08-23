function postData(fieldName, fieldValue, optionalTarget) {
    var $form = $('<form method="POST"><input name="' + fieldName + '" type="hidden"></form>');
    $form.find('input').val(fieldValue);
    if (optionalTarget) {
        $form.prop('action', optionalTarget);
    }
    $(document.body).append($form);
    $form.submit();
    $form.remove();
}