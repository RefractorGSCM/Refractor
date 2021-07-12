CREATE TABLE IF NOT EXISTS ServerGroups(
    ServerID SERIAL NOT NULL,
    GroupID SERIAL NOT NULL,
    AllowOverrides VARCHAR(20) NOT NULL,
    DenyOverrides VARCHAR(20) NOT NULL,

    FOREIGN KEY (ServerID) REFERENCES Servers(ServerID),
    FOREIGN KEY (GroupID) REFERENCES Groups(GroupID)
)