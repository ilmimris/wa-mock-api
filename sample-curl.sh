#!/bin/bash

curl --location 'http://localhost:8080/screenshot' \
--header 'Content-Type: application/json' \
--data '{
    "messages": [
        {
            "timestamp": "2025-05-22T16:48:26.858",
            "sender": "Bot",
            "content": "Hallo Kak, Saya Nia dari Admin SiCepat ingin melakukan konfirmasi atas pengiriman paket dengan detail sebagai berikut:\n\nNo Resi : 016005514153\nPengirim: Fits.ID\nPenerima: Mila Palastri ( Tukang Jahit )\nNo Tlp penerima: +6285642856762\nNilai COD: 0\nIsi Barang: officia. Alat menjahit\nAlamat Pengiriman: Gg. Raya Setiabudhi No. 7Pekanbaru, Kepulauan Riau 43572\n\n*Percobaan pengiriman gagal karena informasi alamat kurang jelas*. Apakah saat ini kakak sudah menerima paket tersebut?\n\nMohon balas pesan ini dengan:\n \"Ya\" jika sudah menerima \n\"Belum\" jika belum menerima \n\"Batalkan\" jika tidak merasa pesan ",
            "recipient_name": "Mila Palastri ( Tukang Jahit )",
            "recipient_phone": "+6285642856762"
        },
        {
            "timestamp": "2025-05-22T16:49:54.293",
            "sender": "Customer",
            "content": "belum",
            "recipient_name": "Mila Palastri ( Tukang Jahit )",
            "recipient_phone": "+6285642856762"
        },
        {
            "timestamp": "2025-05-22T16:49:56.108",
            "sender": "Bot",
            "content": "Baik, terima kasih atas konfirmasinya. \nKami akan mengatur ulang kurir untuk melanjutkan proses pengiriman.\nMohon balas pesan ini dengan menginformasikan:  \n\n*Alamat Lengkap Pengiriman Ulang:*",
            "recipient_name": "Mila Palastri ( Tukang Jahit )",
            "recipient_phone": "+6285642856762"
        },
        {
            "timestamp": "2025-05-22T16:50:14.609",
            "sender": "Customer",
            "content": "jalan sewawar legok gunung",
            "recipient_name": "Mila Palastri ( Tukang Jahit )",
            "recipient_phone": "+6285642856762"
        },
        {
            "timestamp": "2025-05-22T16:50:17.29",
            "sender": "Bot",
            "content": "Baik, terima kasih atas konfirmasinya. Selanjutnya, mohon balas pesan ini dengan meniginformasikan:\n\n*Tanggal & Jam Pengiriman Ulang:*\n",
            "recipient_name": "Mila Palastri ( Tukang Jahit )",
            "recipient_phone": "+6285642856762"
        },
        {
            "timestamp": "2025-05-22T16:50:33.089",
            "sender": "Customer",
            "content": "besok aja, besok sore ya kalau saya udah pulang",
            "recipient_name": "Mila Palastri ( Tukang Jahit )",
            "recipient_phone": "+6285642856762"
        },
        {
            "timestamp": "2025-05-22T16:50:36.043",
            "sender": "Bot",
            "content": "Baik, mohon untuk menyiapkan dana COD senilai <Rp. 0.00> . Terima kasih dan selamat beraktifitas kembali.",
            "recipient_name": "Mila Palastri ( Tukang Jahit )",
            "recipient_phone": "+6285642856762"
        },
        {
            "timestamp": "2025-05-22T16:50:57.724",
            "sender": "Customer",
            "content": "oke2",
            "recipient_name": "Mila Palastri ( Tukang Jahit )",
            "recipient_phone": "+6285642856762"
        }
    ],
    "options": {
        "headerDisplay": "name"
    }
}'
--output screenshot.png

# Check if the curl command was successful
if [ $? -eq 0 ]; then
    echo "Screenshot saved as screenshot.png"
else
    echo "Failed to take screenshot"
fi