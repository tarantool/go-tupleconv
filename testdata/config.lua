-- Do not set listen for now so connector won't be
-- able to send requests until everything is configured.

box.cfg {
    work_dir = os.getenv("TEST_TNT_WORK_DIR"),
}

box.once('init', function()
    box.schema.user.create('test', { password = 'password' })
    box.schema.user.grant('test', 'execute,read,write', 'universe')

    box.schema.create_space('test_space', {
        format = {
            { name = 'id', type = 'unsigned' },
            { name = 'boolean', type = 'boolean' },
            { name = 'number', type = 'number' },
            { name = 'decimal', type = 'decimal' },
            { name = 'datetime', type = 'datetime' },
            { name = 'interval', type = 'interval', is_nullable = true },
            { name = 'string', type = 'string', is_nullable = true },
            { name = 'uuid', type = 'uuid', is_nullable = true },
            { name = 'array', type = 'array' },
            { name = 'any', type = 'any' },
            { name = 'scalar', type = 'scalar', is_nullable = true },
        }
    })

    box.space.test_space:create_index('primary')

    box.schema.create_space('finances', {
        format = {
            { name = 'id', type = 'unsigned' },
            { name = 'money', type = 'number' },
            { name = 'date', type = 'datetime' }
        }
    })

    box.space.finances:create_index('primary')

    datetime = require('datetime')
    box.space.finances:insert { 1, 14.15, datetime.new({ year = 2023, month = 8, day = 30, hour = 12, min = 13 }) }
    box.space.finances:insert { 2, 19.3e4, datetime.new({ year = 2023, month = 8, day = 31, tz = 'Europe/Paris' }) }
    box.space.finances:insert { 3, -111111, datetime.new({ year = 2023, month = 9, day = 2, min = 1, tzoffset = 240 }) }
end)

function get_test_space_fmt()
    return box.space.test_space:format()
end

-- Set listen only when every other thing is configured.
box.cfg {
    listen = os.getenv("TEST_TNT_LISTEN"),
}
