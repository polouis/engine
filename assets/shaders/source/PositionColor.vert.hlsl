cbuffer UniformBlock : register(b0, space1)
{
    float2 offset;
};

struct VSInput
{
    float2 a_position : POSITION;
    float4 a_color    : COLOR;
};

struct VSOutput
{
    float4 position : SV_POSITION;
    float4 color    : COLOR;
};

VSOutput main(VSInput input)
{
    VSOutput output;
    output.position = float4(input.a_position + offset, 0.0, 1.0);
    output.color = input.a_color;
    return output;
}