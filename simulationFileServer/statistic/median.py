import pandas as pd
import os
from glob import glob

# Đường dẫn tới thư mục chứa các file CSV
csv_folder_path = '../'

# Lấy danh sách tất cả các file CSV trong thư mục
csv_files = glob(os.path.join(csv_folder_path, '*.csv'))

# Dictionary để lưu trữ median của từng scheduler cho từng scenario
medians = {}

# Đọc và tính toán median cho từng file CSV
for file in csv_files:
    # Lấy tên file để xác định kịch bản và scheduler
    filename = os.path.basename(file)
    parts = filename.rsplit('-', 1)
    scenario = parts[0]
    scheduler = parts[1].split('.')[0]
    
    # Đọc file CSV
    df = pd.read_csv(file, header=None)
    
    # Tính toán median
    median_value = df.median().values[0]
    
    # Lưu kết quả vào dictionary
    if scenario not in medians:
        medians[scenario] = {'scenario': scenario}
    medians[scenario][scheduler] = median_value

# Chuyển đổi dictionary thành DataFrame
median_df = pd.DataFrame.from_dict(medians, orient='index')

# Sắp xếp các cột theo đúng thứ tự scenario và scheduler
median_df = median_df.sort_values(by='scenario')

# Lưu DataFrame vào file Excel
output_path = './median_results_corrected.xlsx'
median_df.to_excel(output_path, index=False)

output_path
