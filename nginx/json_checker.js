function check_json(r) {
    if (r.method !== 'POST') {
        return;
    }

    var request_body = r.requestBody;
    try {
        JSON.parse(request_body);
    } catch (e) {
        r.error_log("Malformed JSON in POST request body: " + e);
        r.return(400);
    }
}