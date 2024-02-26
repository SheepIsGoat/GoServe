\c server_db

SELECT
	a.id "account_id",
	i.id "invoice_id",
	a.avatar,
	a.name,
	a.title,
	i.amount,
	i.status,
	i.date
FROM "SampleAccounts" a JOIN "SampleInvoices" i 
	ON a.id = i.account_id 