Ergani Go SDKA Go SDK for interacting with the ERGANI API. This SDK is a port of the official Python SDK.Installationgo get github.com/your-username/ergani-go-sdk
(Note: Replace your-username with your actual GitHub username or the appropriate path once you host it.)UsageFirst, you need to import the ergani package.import "github.com/your-username/ergani-go-sdk/ergani"
InitializationInitialize the client with the desired environment (Development or Production) and your Special-Case-Code.client := ergani.NewClient(ergani.Development, "YOUR_SPECIAL_CASE_CODE")
BeneficiariesGet all beneficiariesbeneficiaries, err := client.GetBeneficiaries()
if err != nil {
    // handle error
}
// do something with beneficiaries
Get a single beneficiarybeneficiary, err := client.GetBeneficiary("BENEFICIARY_AFM")
if err != nil {
    // handle error
}
// do something with beneficiary
Create or update a beneficiarynewBeneficiary := ergani.Beneficiary{
    Parartima: "1",
    Afm:       "123456789",
    Surname:   "Test",
    Name:      "Beneficiary",
    // ... fill in other required fields
}
response, err := client.SetBeneficiary(newBeneficiary)
if err != nil {
    // handle error
}
// do something with response
EmployeesGet all employeesemployees, err := client.GetEmployees()
if err != nil {
    // handle error
}
// do something with employees
Get a single employeeemployee, err := client.GetEmployee("EMPLOYEE_AFM")
if err != nil {
    // handle error
}
// do something with employee
Create or update an employeenewEmployee := ergani.Employee{
    Parartima: "1",
    Afm:       "987654321",
    Surname:   "Test",
    Name:      "Employee",
    // ... fill in other required fields
}
response, err := client.SetEmployee(newEmployee)
if err != nil {
    // handle error
}
// do something with response
AnnouncementsGet all announcementsannouncements, err := client.GetAnnouncements()
if err != nil {
    // handle error
}
// do something with announcements
Cancel an announcementcancelPayload := ergani.CancelAnnouncement{
    Protocol: "PROTOCOL_TO_CANCEL",
    Comments: "Cancellation reason.",
}
response, err := client.CancelAnnouncement(cancelPayload)
if err != nil {
    // handle error
}
// do something with response
