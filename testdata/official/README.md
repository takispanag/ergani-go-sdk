# Official reference payloads

The JSON files in this directory are the reference payloads the Greek
Ministry of Labour publishes for the Ergani API, taken from the
`HELP_FILES_SAMPLES_trial2.zip` archive of the interoperability guide.

They are the ground truth of the wire format. The test suite builds the same
declarations with the SDK and asserts that the payload it produces matches
them, field for field.

The two work card samples are the exception: `Card.json` carries a trailing
comma, which is not valid JSON, and `f_aitiologia` is an empty string where
the API documents null. Both are as published.

The files are not covered by the licence of this repository. They belong to the
Ministry of Labour and are reproduced here so that the format they describe
can be tested against.
