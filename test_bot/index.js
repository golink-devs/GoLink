require('dotenv').config();
const { Client, GatewayIntentBits } = require('discord.js');
const { Shoukaku, Connectors } = require('shoukaku');

const Nodes = [{
    name: process.env.GOLINK_NODE_NAME || 'GoLink',
    url: process.env.GOLINK_NODE_URL || 'localhost:2333',
    auth: process.env.GOLINK_NODE_AUTH || 'youshallnotpass',
    secure: process.env.GOLINK_NODE_SECURE === 'true'
}];

const client = new Client({
    intents: [
        GatewayIntentBits.Guilds,
        GatewayIntentBits.GuildMessages,
        GatewayIntentBits.MessageContent,
        GatewayIntentBits.GuildVoiceStates
    ]
});

const shoukaku = new Shoukaku(new Connectors.DiscordJS(client), Nodes);

shoukaku.on('ready', (name) => console.log(`Node ${name} is ready!`));
shoukaku.on('error', (name, error) => console.error(`Node ${name} had an error:`, error));

client.on('messageCreate', async (message) => {
    if (message.author.bot || !message.content.startsWith('!play')) return;

    const query = message.content.slice(6).trim();
    if (!query) return message.reply('Please provide a search query or URL!');

    const node = shoukaku.options.nodeResolver(shoukaku.nodes);
    const result = await node.rest.resolve(`ytsearch:${query}`);

    if (!result || !result.data || result.loadType === 'empty' || result.loadType === 'error') {
        return message.reply('No results found!');
    }

    const track = result.loadType === 'search' ? result.data[0] : result.data;
    const voiceChannel = message.member.voice.channel;
    if (!voiceChannel) return message.reply('You must be in a voice channel!');

    const player = await node.joinChannel({
        guildId: message.guild.id,
        channelId: voiceChannel.id,
        shardId: 0
    });

    await player.playTrack({ track: track.encoded });
    message.reply(`Now playing: **${track.info.title}**`);
});

client.on('ready', () => console.log(`Bot ${client.user.tag} is ready!`));

const token = process.env.DISCORD_TOKEN;
if (!token) {
    console.error('DISCORD_TOKEN environment variable not set!');
    process.exit(1);
}

client.login(token);
