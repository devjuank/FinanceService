import json
import os
import re

REQUIRED_FIELDS = {
    "transaction_id": str,
    "source": str,
    "account": str,
    "date": str,
    "amount": (int, float),
    "currency": str,
    "description": str,
    "direction": str,
    "merchant": (str, type(None)),
    "category": (str, type(None)),
    "subcategory": (str, type(None)),
    "balance": (int, float, type(None)),
    "is_transfer": bool,
    "is_fee": bool,
    "is_tax": bool
}

def validate_iso_date(date_str):
    return bool(re.match(r'^\d{4}-\d{2}-\d{2}$', date_str))

def validate_transaction(tx, index, file_path):
    errors = []
    
    # Check for missing fields
    for field, expected_type in REQUIRED_FIELDS.items():
        if field not in tx:
            errors.append(f"Missing field: {field}")
            continue
            
        # Check type
        if not isinstance(tx[field], expected_type):
            errors.append(f"Invalid type for {field}: expected {expected_type}, got {type(tx[field])}")
            
    # Specific validations
    if "date" in tx and not validate_iso_date(tx["date"]):
        errors.append(f"Invalid date format: {tx['date']} (expected YYYY-MM-DD)")
        
    if "direction" in tx and tx["direction"] not in ["debit", "credit"]:
        errors.append(f"Invalid direction: {tx['direction']} (expected debit or credit)")
        
    if "transaction_id" in tx and not re.match(r'^[a-f0-9]{64}$', tx["transaction_id"]):
        errors.append(f"Invalid transaction_id format (expected 64-char hex hash)")

    return errors

def validate_file(file_path):
    print(f"Validating {file_path}...")
    try:
        with open(file_path, 'r') as f:
            data = json.load(f)
            
        if not isinstance(data, list):
            print(f"  [ERROR] Root element must be an array.")
            return False
            
        total_errors = 0
        for idx, tx in enumerate(data):
            errors = validate_transaction(tx, idx, file_path)
            if errors:
                print(f"  [ERROR] Transaction at index {idx}:")
                for err in errors:
                    print(f"    - {err}")
                total_errors += len(errors)
                
        if total_errors == 0:
            print(f"  [SUCCESS] {len(data)} transactions validated successfully.")
            return True
        else:
            print(f"  [FAILED] {total_errors} errors found.")
            return False
            
    except Exception as e:
        print(f"  [ERROR] Could not read file: {e}")
        return False

def main():
    target_dir = "/Users/juank/Documents/Cuentas/DatosClasificados"
    if not os.path.exists(target_dir):
        print(f"Directory not found: {target_dir}")
        return

    files_to_validate = [f for f in os.listdir(target_dir) if f.endswith('.json') and f != 'consolidated_transactions.json']
    
    all_ok = True
    for file_name in files_to_validate:
        if not validate_file(os.path.join(target_dir, file_name)):
            all_ok = False
            
    if all_ok:
        print("\nAll files are valid!")
    else:
        print("\nSome files failed validation.")

if __name__ == "__main__":
    main()
