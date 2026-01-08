import os
import json
import hashlib
import re
import pandas as pd
from pypdf import PdfReader
from datetime import datetime

class FinancialNormalizer:
    def __init__(self, output_dir):
        self.output_dir = output_dir
        if not os.path.exists(output_dir):
            os.makedirs(output_dir)

    def generate_id(self, source, account, date, amount, raw_description):
        payload = f"{source}{account}{date}{amount}{raw_description}"
        return hashlib.sha256(payload.encode()).hexdigest()

    def clean_amount(self, amount_str):
        if not amount_str or amount_str == '-' or str(amount_str).lower() == 'nan':
            return 0.0
        # Remove currency symbols and other non-numeric chars except . , -
        clean = re.sub(r'[^\d,\.\-]', '', str(amount_str)).strip()
        if not clean: return 0.0
        
        # If there's a comma and a dot, it's likely ARG format: 1.234,56
        if ',' in clean and '.' in clean:
            clean = clean.replace('.', '').replace(',', '.')
        # If there's only a comma, it's the decimal separator: 1234,56
        elif ',' in clean:
            clean = clean.replace(',', '.')
        
        try:
            return float(clean)
        except ValueError:
            return 0.0

    def infer_category(self, description):
        desc = description.lower()
        if any(w in desc for w in ['iva', 'percepción', 'ganancias', 'tax', 'impuesto', 'sircreb', 'arca', 'afip']):
            return 'impuestos', 'impuestos y contribuciones'
        if any(w in desc for w in ['netflix', 'spotify', 'youtube', 'primevideo', 'disney']):
            return 'entretenimiento', 'servicios digitales'
        if any(w in desc for w in ['pedidosya', 'rappi', 'mcdonalds', 'burger']):
            return 'comida', 'delivery'
        if any(w in desc for w in ['metrogas', 'aysa', 'edenor', 'edesur', 'personal flow', 'claro']):
            return 'servicios', 'hogar'
        if any(w in desc for w in ['intereses pagados']):
            return 'financiero', 'intereses'
        if any(w in desc for w in ['reintegro promoción']):
            return 'ingresos', 'reintegros'
        return None, None

    def normalize_brubank_pdf(self, pdf_path):
        reader = PdfReader(pdf_path)
        lines = []
        for page in reader.pages:
            lines.extend(page.extract_text().split('\n'))

        transactions = []
        i = 0
        while i < len(lines):
            # Brubank transactions usually follow a date pattern: DD-MM-YY
            if re.match(r'^\d{2}-\d{2}-\d{2}$', lines[i]):
                try:
                    date_raw = lines[i]
                    date_iso = datetime.strptime(date_raw, '%d-%m-%y').strftime('%Y-%m-%d')
                    ref = lines[i+1]
                    description = lines[i+2]
                    
                    # Sometimes descriptions spans multiple lines (handled by logic or ignored if noise)
                    # We need to find Debit, Credit, Saldo
                    # Based on observation, they follow Ref.
                    debit_str = lines[i+3]
                    credit_str = lines[i+4]
                    balance_str = lines[i+5]
                    
                    debit = self.clean_amount(debit_str)
                    credit = self.clean_amount(credit_str)
                    balance = self.clean_amount(balance_str)
                    
                    amount = credit - debit
                    direction = 'credit' if amount > 0 else 'debit'
                    
                    is_tax = any(w in description.lower() for w in ['iva', 'percepción', 'ganancias', 'impuesto', 'sircreb', 'arca', 'afip'])
                    is_transfer = any(w in description.lower() for w in ['cuenta tuya', 'transferencia'])
                    is_fee = any(w in description.lower() for w in ['comisión', 'reimpresión', 'intereses pagados'])
                    
                    cat, subcat = self.infer_category(description)
                    
                    # Extract merchant (very basic logic)
                    merchant = None
                    if not is_tax and not is_transfer and not is_fee:
                        merchant = description.split(' ')[0] if ' ' in description else description

                    tx = {
                        "transaction_id": self.generate_id("brubank", "caja_ahorro_pesos", date_iso, amount, description),
                        "source": "brubank",
                        "account": "caja_ahorro_pesos",
                        "date": date_iso,
                        "amount": round(amount, 2),
                        "currency": "ARS",
                        "description": description.strip(),
                        "direction": direction,
                        "merchant": merchant,
                        "category": cat,
                        "subcategory": subcat,
                        "balance": balance,
                        "is_transfer": is_transfer,
                        "is_fee": is_fee,
                        "is_tax": is_tax
                    }
                    transactions.append(tx)
                    i += 6 # Skip to next potential block
                    continue
                except Exception as e:
                    pass
            i += 1
        return transactions

    def normalize_mercadopago_csv(self, csv_path):
        try:
            with open(csv_path, 'r', encoding='utf-8') as f:
                lines = f.readlines()
            
            # Find the line that starts with RELEASE_DATE
            header_idx = -1
            sep = ','
            for idx, line in enumerate(lines):
                if 'RELEASE_DATE' in line:
                    header_idx = idx
                    sep = ';' if ';' in line else ','
                    break
            
            if header_idx == -1:
                return []
                
            # Read from the header line onwards
            df = pd.read_csv(csv_path, sep=sep, skiprows=header_idx)
            
            transactions = []
            col_map = {
                'date': 'RELEASE_DATE',
                'description': 'TRANSACTION_TYPE',
                'amount': 'TRANSACTION_NET_AMOUNT',
                'balance': 'PARTIAL_BALANCE'
            }
            
            if col_map['date'] not in df.columns or col_map['amount'] not in df.columns:
                return []

            for _, row in df.iterrows():
                try:
                    date_val = str(row[col_map['date']])
                    if not date_val or date_val == 'nan' or date_val.strip() == '': continue
                    
                    # Some lines might be empty or footer
                    if not re.match(r'\d{2}-\d{2}-\d{4}', date_val):
                        continue

                    date_iso = datetime.strptime(date_val, '%d-%m-%Y').strftime('%Y-%m-%d')
                    
                    amount_raw = str(row[col_map['amount']])
                    amount = self.clean_amount(amount_raw)
                    
                    description = str(row[col_map['description']])
                    balance_raw = str(row[col_map['balance']]) if col_map['balance'] in row else None
                    balance = self.clean_amount(balance_raw) if balance_raw else None
                    
                    direction = 'credit' if amount > 0 else 'debit'
                    is_tax = any(w in description.lower() for w in ['percepción', 'iva', 'impuesto', 'arca'])
                    is_transfer = any(w in description.lower() for w in ['transferencia', 'enviaste', 'recibiste', 'de una cuenta tuya', 'a una cuenta tuya'])
                    is_fee = 'comisión' in description.lower()
                    
                    cat, subcat = self.infer_category(description)
                    
                    merchant = None
                    if 'Pago ' in description:
                        merchant = description.replace('Pago ', '').split(' ')[0]
                    elif 'Transferencia ' in description:
                        merchant = description.replace('Transferencia ', '').replace('enviada ', '').replace('recibida ', '')

                    tx = {
                        "transaction_id": self.generate_id("mercadopago", "cuenta_digital", date_iso, amount, description),
                        "source": "mercadopago",
                        "account": "cuenta_digital",
                        "date": date_iso,
                        "amount": round(amount, 2),
                        "currency": "ARS",
                        "description": description.strip(),
                        "direction": direction,
                        "merchant": merchant,
                        "category": cat,
                        "subcategory": subcat,
                        "balance": balance,
                        "is_transfer": is_transfer,
                        "is_fee": is_fee,
                        "is_tax": is_tax
                    }
                    transactions.append(tx)
                except Exception as e:
                    continue
            return transactions
        except Exception as e:
            print(f"Error processing {csv_path}: {e}")
            return []

    def normalize_deel_csv(self, csv_path):
        try:
            df = pd.read_csv(csv_path)
            
            # Deel CSV columns: ID, Date Requested, Transaction Status, Transaction Type, Currency, Transaction Amount, etc.
            if 'Date Requested' not in df.columns or 'Transaction Amount' not in df.columns:
                return []

            transactions = []
            for _, row in df.iterrows():
                try:
                    status = str(row['Transaction Status']).lower()
                    if status != 'completed':
                        continue

                    date_val = str(row['Date Requested'])
                    # Format: YYYY-MM-DD HH:MM:SS
                    date_iso = date_val.split(' ')[0]
                    
                    amount = float(row['Transaction Amount'])
                    currency = str(row['Currency'])
                    tx_type = str(row['Transaction Type'])
                    
                    client = str(row['Client']) if not pd.isna(row['Client']) else ""
                    contract = str(row['Contract Name']) if not pd.isna(row['Contract Name']) else ""
                    description = f"{tx_type}: {client} {contract}".strip()
                    
                    direction = 'credit' if amount > 0 else 'debit'
                    is_tax = any(w in description.lower() for w in ['tax', 'fee']) # Deel has provider fees
                    is_transfer = tx_type in ['withdrawal', 'deel_card_withdrawal']
                    is_fee = 'fee' in description.lower() or tx_type == 'provider_fee'
                    
                    cat, subcat = self.infer_category(description)
                    if not cat and tx_type == 'client_payment':
                        cat, subcat = 'ingresos', 'sueldo'

                    tx = {
                        "transaction_id": str(row['ID']), # Deel already provides a UUID
                        "source": "deel",
                        "account": "balance_usd",
                        "date": date_iso,
                        "amount": round(amount, 2),
                        "currency": currency,
                        "description": description,
                        "direction": direction,
                        "merchant": client if client != "" else None,
                        "category": cat,
                        "subcategory": subcat,
                        "balance": None, # Deel CSV doesn't seem to have running balance
                        "is_transfer": is_transfer,
                        "is_fee": is_fee,
                        "is_tax": is_tax
                    }
                    transactions.append(tx)
                except Exception as e:
                    continue
            return transactions
        except Exception as e:
            print(f"Error processing {csv_path}: {e}")
            return []

    def normalize_santander_xlsx(self, xlsx_path):
        try:
            # We observed header is at row 12 (skiprows=12)
            df = pd.read_excel(xlsx_path, skiprows=12)
            
            # Columns are likely unnamed. Based on observation:
            # idx 1: Date, idx 3: Description, idx 6: Amount, idx 7: Balance
            transactions = []
            for _, row in df.iterrows():
                try:
                    date_val = str(row.iloc[1])
                    if not re.match(r'\d{2}/\d{2}/\d{4}', date_val):
                        continue
                    
                    date_iso = datetime.strptime(date_val, '%d/%m/%Y').strftime('%Y-%m-%d')
                    description = str(row.iloc[3]).strip()
                    amount = self.clean_amount(str(row.iloc[6]))
                    balance = self.clean_amount(str(row.iloc[7]))
                    
                    if amount == 0: continue

                    direction = 'credit' if amount > 0 else 'debit'
                    is_tax = any(w in description.lower() for w in ['impuesto', 'iva', 'percepción', 'sircreb', 'db.rg'])
                    is_transfer = 'transferencia' in description.lower()
                    is_fee = any(w in description.lower() for w in ['comision', 'cargo', 'interes'])
                    
                    cat, subcat = self.infer_category(description)

                    tx = {
                        "transaction_id": self.generate_id("santander", "caja_ahorro_pesos", date_iso, amount, description),
                        "source": "santander",
                        "account": "caja_ahorro_pesos",
                        "date": date_iso,
                        "amount": round(amount, 2),
                        "currency": "ARS",
                        "description": description,
                        "direction": direction,
                        "merchant": None,
                        "category": cat,
                        "subcategory": subcat,
                        "balance": balance,
                        "is_transfer": is_transfer,
                        "is_fee": is_fee,
                        "is_tax": is_tax
                    }
                    transactions.append(tx)
                except:
                    continue
            return transactions
        except Exception as e:
            print(f"Error processing {xlsx_path}: {e}")
            return []

    def normalize_santander_visa_pdf(self, pdf_path):
        try:
            reader = PdfReader(pdf_path)
            transactions = []
            
            # Regex for Santander/Visa transactions:
            # Date format: 25 Agosto 02 or 02 Mayo ...
            # Then some code like 006483 K
            # Then Description
            # Then Amount (sometimes with minus sign at the end or start)
            
            # Pattern: Day Month Day Description Amount
            # Months are usually in Spanish
            months_map = {
                'Enero': '01', 'Febrero': '02', 'Marzo': '03', 'Abril': '04', 'Mayo': '05', 'Junio': '06',
                'Julio': '07', 'Agosto': '08', 'Setiembre': '09', 'Septiembre': '09', 'Octubre': '10', 'Noviembre': '11', 'Diciembre': '12',
                'Ene': '01', 'Feb': '02', 'Mar': '03', 'Abr': '04', 'May': '05', 'Jun': '06',
                'Jul': '07', 'Ago': '08', 'Set': '09', 'Sep': '09', 'Oct': '10', 'Nov': '11', 'Dic': '12'
            }
            
            text = ""
            for page in reader.pages:
                text += page.extract_text() + "\n"
            
            lines = text.split('\n')
            
            # Find the year in the header (CIERRE 28 Ago 25)
            year_match = re.search(r'CIERRE\s+\d{2}\s+\w{3}\s+(\d{2})', text)
            year_prefix = "20" + year_match.group(1) if year_match else "2025"

            for line in lines:
                # Example: 02 Agosto 02 006483 K PEDIDOSYA 27.269,00
                # or: 25 Agosto 05 SU PAGO EN PESOS 749.728,51-
                match = re.search(r'(\d{2})\s+([a-zA-Z]{3,10})\s+(\d{2})\s+.*?\s+(.*?)\s+([\d\.\,]+-?)(\s+[\d\.\,]+-?)?$', line)
                if match:
                    day_resumen = match.group(1)
                    month_str = match.group(2).capitalize()
                    day_tx = match.group(3)
                    description = match.group(4).strip()
                    amount_str = match.group(5).strip()
                    # group 6 might be another currency or U$S amount in the same row
                    
                    if month_str not in months_map: continue
                    
                    month = months_map[month_str]
                    # We use year from closing if tx month is same or before, otherwise year-1 (if closing is Jan and tx is Dec)
                    # For simplicity, we use year_prefix
                    date_iso = f"{year_prefix}-{month}-{day_tx}"
                    
                    # Clean amount
                    is_negative = False
                    if amount_str.endswith('-'):
                        is_negative = True
                        amount_str = amount_str[:-1]
                    
                    amount = self.clean_amount(amount_str)
                    
                    # On Credit Card statements, payments/credits are usually positive in the file (but reduced balance)
                    # and purchases are presented as items. 
                    # Santander Visa specifically: purchases have NO minus sign, payments have minus sign.
                    # so if is_negative is True -> it's a credit (payment), if False -> it's a debit (purchase)
                    if is_negative:
                        amount = abs(amount)
                        direction = 'credit'
                    else:
                        amount = -abs(amount)
                        direction = 'debit'

                    if amount == 0: continue

                    is_tax = any(w in description.lower() for w in ['impuesto', 'iva', 'percepción', 'db.rg'])
                    is_transfer = 'SU PAGO' in description or 'PAGO EN' in description
                    is_fee = any(w in description.lower() for w in ['comision', 'cargo', 'interes'])
                    
                    cat, subcat = self.infer_category(description)

                    tx = {
                        "transaction_id": self.generate_id("santander", "credito_visa", date_iso, amount, description),
                        "source": "santander",
                        "account": "credito_visa",
                        "date": date_iso,
                        "amount": round(amount, 2),
                        "currency": "ARS", # Could be USD if extracted from the second amount column, but standardizing for now
                        "description": description,
                        "direction": direction,
                        "merchant": description.split(' ')[0],
                        "category": cat,
                        "subcategory": subcat,
                        "balance": None,
                        "is_transfer": is_transfer,
                        "is_fee": is_fee,
                        "is_tax": is_tax
                    }
                    transactions.append(tx)
                
            return transactions
        except Exception as e:
            print(f"Error processing {pdf_path}: {e}")
            return []

    def run(self):
        all_transactions = []

        # Process Brubank
        brubank_dir = "/Users/juank/Documents/Cuentas/Bancos/Brubank"
        all_brubank = []
        for f in os.listdir(brubank_dir):
            if f.endswith('.pdf'):
                all_brubank.extend(self.normalize_brubank_pdf(os.path.join(brubank_dir, f)))
        
        if all_brubank:
            with open(os.path.join(self.output_dir, 'brubank.json'), 'w') as f:
                json.dump(all_brubank, f, indent=2)
            pd.DataFrame(all_brubank).to_csv(os.path.join(self.output_dir, 'brubank.csv'), index=False)
            all_transactions.extend(all_brubank)

        # Process MercadoPago
        mp_dir = "/Users/juank/Documents/Cuentas/Bancos/MercadoPago"
        all_mp = []
        for f in os.listdir(mp_dir):
            if f.endswith('.csv'):
                all_mp.extend(self.normalize_mercadopago_csv(os.path.join(mp_dir, f)))
        
        if all_mp:
            with open(os.path.join(self.output_dir, 'mercadopago.json'), 'w') as f:
                json.dump(all_mp, f, indent=2)
            pd.DataFrame(all_mp).to_csv(os.path.join(self.output_dir, 'mercadopago.csv'), index=False)
            all_transactions.extend(all_mp)

        # Process Deel
        deel_dir = "/Users/juank/Documents/Cuentas/Bancos/Deel"
        all_deel = []
        for f in os.listdir(deel_dir):
            if f.endswith('.csv'):
                all_deel.extend(self.normalize_deel_csv(os.path.join(deel_dir, f)))
        
        if all_deel:
            with open(os.path.join(self.output_dir, 'deel.json'), 'w') as f:
                json.dump(all_deel, f, indent=2)
            pd.DataFrame(all_deel).to_csv(os.path.join(self.output_dir, 'deel.csv'), index=False)
            all_transactions.extend(all_deel)

        # Process Santander
        santander_dir = "/Users/juank/Documents/Cuentas/Bancos/santander"
        all_santander = []
        for f in os.listdir(santander_dir):
            if f.endswith('.xlsx'):
                all_santander.extend(self.normalize_santander_xlsx(os.path.join(santander_dir, f)))
        
        tarjeta_dir = os.path.join(santander_dir, "Tarjeta")
        if os.path.exists(tarjeta_dir):
            for f in os.listdir(tarjeta_dir):
                if f.endswith('.pdf'):
                    all_santander.extend(self.normalize_santander_visa_pdf(os.path.join(tarjeta_dir, f)))
        
        if all_santander:
            with open(os.path.join(self.output_dir, 'santander.json'), 'w') as f:
                json.dump(all_santander, f, indent=2)
            pd.DataFrame(all_santander).to_csv(os.path.join(self.output_dir, 'santander.csv'), index=False)
            all_transactions.extend(all_santander)

        # Consolidated Result
        if all_transactions:
            pd.DataFrame(all_transactions).sort_values('date', ascending=False).to_csv(
                os.path.join(self.output_dir, 'consolidated_transactions.csv'), index=False)

if __name__ == "__main__":
    normalizer = FinancialNormalizer("/Users/juank/Documents/Cuentas/DatosClasificados")
    normalizer.run()
