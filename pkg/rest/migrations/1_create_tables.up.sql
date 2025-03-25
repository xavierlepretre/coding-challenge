CREATE TABLE Bill (
    CustomerId TEXT NOT NULL,
    Id TEXT NOT NULL,
    CurrencyCode TEXT NOT NULL,
    Status INT NOT NULL DEFAULT 0,
    LineItemCount BIGINT NOT NULL DEFAULT 0,
    TotalAmount BIGINT NOT NULL DEFAULT 0,
    TotalOk BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (CustomerId, Id)
);

CREATE TABLE LineItem (
    CustomerId TEXT NOT NULL,
    BillId TEXT NOT NULL,
    Id TEXT NOT NULL,
    Description TEXT NOT NULL,
    Amount BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (CustomerId, BillId, Id)
);