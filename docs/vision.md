## Problems with v1
- CLI interface was tedious to use, had to pass a lot of arguments on the command line every time
- no way to filter based on wildcards, if a transaction from the same company at a different store had one character difference in the transaction description, it would require manual user categorization
- no way to edit a transaction that should end up in a different category for only a single month, all or nothing schema
- output was stored in individual files, no nice way to load information into the tool and have it persist
- no good way to configure tool locally
- relied on specific "parser types" for each CSV file, didn't actually read CSV header and then intelligently normalize the data

## Key Functionality

### Parse Bank Transactions
- parse CSV files containing bank transactions
- should use CSV header to determine what fields matter, normalized against an internal set of fields

### Categorize Bank Transactions
- categorize transactions based on user input
- user should define categories that specify the "budget", for example:
```
{
    "<category name>": {
        "description": "describe what goes in this category",
        "budget": "monthly amount budgeted for this category",
    }
}
```
- shouldn't prompt a user to categorize a transaction from the same place twice
- store all parsed transactions so that they can be used to output summary

### Output Summary CSV
- create a summary CSV file of all the transactions for a given month, including category