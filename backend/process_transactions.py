import pandas as pd
import json
import os
from datetime import datetime, timedelta

def process_transactions(input_file, output_file, rules_file):
    if not os.path.exists(input_file):
        print(f"Error: {input_file} not found.")
        return

    # 1. Load data
    df = pd.read_csv(input_file)
    print(f"Original records: {len(df)}")

    # 2. Deduplicate by transaction_id
    df = df.drop_duplicates(subset=['transaction_id'], keep='first')
    print(f"Records after deduplication: {len(df)}")

    # 3. Load rules
    with open(rules_file, 'r') as f:
        rules = json.load(f)

    # 4. Improved Categorization
    def categorize(row):
        desc = str(row['description']).lower()
        merchant = str(row['merchant']).lower() if not pd.isna(row['merchant']) else ""
        
        # Check merchant rules
        for m_key, m_val in rules['merchants'].items():
            if m_key.lower() in merchant or m_key.lower() in desc:
                return m_val['category'], m_val['subcategory']
        
        # Check description keywords
        for d_key, d_val in rules['description_keywords'].items():
            if d_key.lower() in desc:
                return d_val['category'], d_val['subcategory']
        
        return row['category'], row['subcategory']

    new_cats = df.apply(categorize, axis=1)
    df['category'] = [c[0] for c in new_cats]
    df['subcategory'] = [c[1] for c in new_cats]

    # 5. Transfer Neutralization
    df['date'] = pd.to_datetime(df['date'])
    df['neutralized'] = False
    
    transfers = df[df['is_transfer'] == True].copy()
    debits = transfers[df['amount'] < 0].copy()
    credits = transfers[df['amount'] > 0].copy()

    neutralized_ids = set()

    for idx_d, row_d in debits.iterrows():
        if row_d['transaction_id'] in neutralized_ids:
            continue
            
        target_amount = abs(row_d['amount'])
        # Look for a credit with similar amount (+/- 0.5% for potential small fees or rounding)
        mask = (credits['amount'] >= target_amount * 0.995) & \
               (credits['amount'] <= target_amount * 1.005) & \
               (credits['date'] >= row_d['date'] - timedelta(days=1)) & \
               (credits['date'] <= row_d['date'] + timedelta(days=3)) & \
               (~credits['transaction_id'].isin(neutralized_ids))
        
        match = credits[mask]
        if not match.empty:
            match_idx = match.index[0]
            neutralized_ids.add(row_d['transaction_id'])
            neutralized_ids.add(match.iloc[0]['transaction_id'])
            
            # Mark as internal transfer
            df.at[idx_d, 'category'] = 'transferencia_interna'
            df.at[idx_d, 'neutralized'] = True
            df.at[match_idx, 'category'] = 'transferencia_interna'
            df.at[match_idx, 'neutralized'] = True

    print(f"Neutralized {len(neutralized_ids)} transactions ({len(neutralized_ids)//2} pairs).")

    # 6. Save results
    df['date'] = df['date'].dt.strftime('%Y-%m-%d')
    df.to_csv(output_file, index=False)
    print(f"Processed file saved to: {output_file}")

if __name__ == "__main__":
    input_f = "/Users/juank/Documents/Cuentas/DatosClasificados/consolidated_transactions.csv"
    output_f = "/Users/juank/Documents/Cuentas/DatosClasificados/processed_transactions.csv"
    rules_f = "/Users/juank/Documents/Cuentas/classification_rules.json"
    process_transactions(input_f, output_f, rules_f)
